package main

import (
	"context"
	"strings"

	"google.golang.org/api/cloudresourcemanager/v2"
)

const systemGSuiteFolderName = "system-gsuite"
const appsScriptFolderName = "apps-script"

func listDescendantFolders(ctx context.Context, parent string) ([]*cloudresourcemanager.Folder, error) {
	svc, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return listDescendantFoldersImpl(ctx, cloudresourcemanager.NewFoldersService(svc), parent)
}

func listDescendantFoldersImpl(ctx context.Context, folderSvc *cloudresourcemanager.FoldersService, parent string) ([]*cloudresourcemanager.Folder, error) {
	var folders []*cloudresourcemanager.Folder
	err := folderSvc.List().Parent(parent).Pages(ctx, func(resp *cloudresourcemanager.ListFoldersResponse) error {
		for i := range resp.Folders {
			folder := resp.Folders[i]
			if folder.DisplayName == systemGSuiteFolderName || folder.DisplayName == appsScriptFolderName {
				continue
			}
			folders = append(folders, folder)
			fs, err := listDescendantFoldersImpl(ctx, folderSvc, folder.Name)
			if err != nil {
				return err
			}
			folders = append(folders, fs...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return folders, nil
}

func formatAncestors(parent string, folders map[string]*cloudresourcemanager.Folder) string {
	return formatFolderPath(getAncestors(folders, parent))
}

func formatFolderPath(path []*cloudresourcemanager.Folder) string {
	var displayNames []string
	for _, folder := range path {
		displayNames = append(displayNames, folder.DisplayName)
	}
	return "/" + strings.Join(displayNames, "/")
}

func getAncestors(folders map[string]*cloudresourcemanager.Folder, parent string) []*cloudresourcemanager.Folder {
	var ancestors []*cloudresourcemanager.Folder
	for parentFolder, ok := folders[parent]; ok; parentFolder, ok = folders[parentFolder.Parent] {
		ancestors = append([]*cloudresourcemanager.Folder{parentFolder}, ancestors...)
	}
	return ancestors
}
