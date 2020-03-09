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

	var parentNames = []string{organizationName}
	for _, folder := range folders {
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
	err = forEachProjectBillingInfo(ctx, projects,
		func(project *cloudresourcemanagerv1.Project, billingInfo *cloudbilling.ProjectBillingInfo) {
			bar.Increment()
			fmt.Fprintf(file, "%v,%v,%v\n", project.ProjectId, formatParent(project.Parent), billingInfo.BillingAccountName)
		})
	bar.Finish()
	if err != nil {
		return err
	}
	return nil
}


