package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cheggaaa/pb/v3"
	"google.golang.org/api/cloudbilling/v1"
	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/cloudresourcemanager/v2"
)

var organization = flag.String("organization", "", "organization ID")
var output = flag.String("output", "", "output file")

func main() {
	flag.Parse()

	if *organization == "" || *output == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := _main(); err != nil {
		log.Fatalln(err)
	}
}

func _main() error {
	ctx := context.Background()
	organizationName := fmt.Sprintf("organizations/%s", *organization)
	folders, err := listDescendantFolders(ctx, organizationName)
	if err != nil {
		return err
	}

	folderMap := make(map[string]*cloudresourcemanager.Folder)
	var parentNames = []string{organizationName}
	for _, folder := range folders {
		folderMap[folder.Name] = folder
		parentNames = append(parentNames, folder.Name)
	}

	projects, err := listChildrenProjectsForEachParents(ctx, parentNames)
	if err != nil {
		return err
	}

	file, err := os.Create(*output)
	if err != nil {
		return err
	}
	defer file.Close()

	bar := pb.StartNew(len(projects))
	fmt.Fprintf(file, "%v,%v,%v\n", "project_id", "billing_account_id", "display_name_path")
	err = forEachProjectBillingInfo(ctx, projects,
		func(project *cloudresourcemanagerv1.Project, billingInfo *cloudbilling.ProjectBillingInfo) {
			bar.Increment()
			fmt.Fprintf(file, "%v,%v,%v\n", project.ProjectId, billingInfo.BillingAccountName, formatAncestors(formatParent(project.Parent), folderMap))
		})
	if err != nil {
		return err
	}
	bar.Finish()
	return nil
}
