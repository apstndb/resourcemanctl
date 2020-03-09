package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"google.golang.org/api/cloudbilling/v1"
	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/cloudresourcemanager/v2"
)

var organization = flag.String("organization", "", "organization ID")

func main() {
	flag.Parse()

	if *organization == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := _main(); err != nil {
		log.Fatalln(err)
	}
}

var folderSvc *cloudresourcemanager.FoldersService
var projectSvc *cloudresourcemanagerv1.ProjectsService
var billingSvc *cloudbilling.ProjectsService

func _main() error {
	ctx := context.Background()
	{
		svc, err := cloudresourcemanagerv1.NewService(ctx)
		if err != nil {
			return err
		}
		projectSvc = cloudresourcemanagerv1.NewProjectsService(svc)
	}
	{
		svc, err := cloudresourcemanager.NewService(ctx)
		if err != nil {
			return err
		}
		folderSvc = cloudresourcemanager.NewFoldersService(svc)
	}
	{
		svc, err := cloudbilling.NewService(ctx)
		if err != nil {
			return err
		}
		billingSvc = cloudbilling.NewProjectsService(svc)
	}

	organizationName := fmt.Sprintf("organizations/%s", *organization)
	folders, err := listDescendantFolders(ctx, organizationName)
	if err != nil {
		return err
	}
	projects, err := listChildrenProjects(ctx, organizationName)
	if err != nil {
		return err
	}
	for i := range folders {
		folder := folders[i]
		// ignore system-gsuite/apps-script
		if folder.DisplayName == "system-gsuite" {
			continue
		}
		pr, err := listChildrenProjects(ctx, folder.Name)
		if err != nil {
			return err
		}
		projects = append(projects, pr...)
	}

	bar := pb.StartNew(len(projects))

	for i := range projects {
		bar.Increment()
		project := projects[i]
		if project.LifecycleState != "ACTIVE" {
			continue
		}
		billingInfo, err := billingSvc.GetBillingInfo("projects/" + project.ProjectId).Do()
		if err != nil {
			return err
		}
		var parentName string = formatParent(project.Parent)
		fmt.Printf("%v,%v,%v\n", project.ProjectId, parentName, billingInfo.BillingAccountName)
	}
	bar.Finish()
	return nil
}

func formatParent(parent *cloudresourcemanagerv1.ResourceId) string {
	if parent == nil {
		return ""
	}
	return fmt.Sprintf("%ss/%s", parent.Type, parent.Id)
}

func listChildrenProjects(ctx context.Context, parent string) ([]*cloudresourcemanagerv1.Project, error) {
	var projects []*cloudresourcemanagerv1.Project
	path := strings.Split(parent, "/")
	err := projectSvc.List().Filter(fmt.Sprintf(`parent.type:%s parent.id:%s`, strings.TrimSuffix(path[0], "s"), path[1])).Pages(ctx, func(resp *cloudresourcemanagerv1.ListProjectsResponse) error {
		for i := range resp.Projects {
			project := resp.Projects[i]
			projects = append(projects, project)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func listDescendantFolders(ctx context.Context, parent string) ([]*cloudresourcemanager.Folder, error) {
	var folders []*cloudresourcemanager.Folder
	err := folderSvc.List().Parent(parent).Pages(ctx, func(resp *cloudresourcemanager.ListFoldersResponse) error {
		for i := range resp.Folders {
			folder := resp.Folders[i]
			folders = append(folders, folder)
			fs, err := listDescendantFolders(ctx, folder.Name)
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
