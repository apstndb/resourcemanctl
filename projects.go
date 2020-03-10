package main

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
)

func formatParent(parent *cloudresourcemanager.ResourceId) string {
	if parent == nil {
		return ""
	}
	return fmt.Sprintf("%ss/%s", parent.Type, parent.Id)
}

func listChildrenProjects(ctx context.Context, parent string) ([]*cloudresourcemanager.Project, error) {
	svc, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return listChildrenProjectsImpl(ctx, cloudresourcemanager.NewProjectsService(svc), parent)
}

func listChildrenProjectsImpl(ctx context.Context, projectSvc *cloudresourcemanager.ProjectsService, parent string) ([]*cloudresourcemanager.Project, error) {
	var projects []*cloudresourcemanager.Project
	err := listChildrenProjectsCall(projectSvc, parent).Pages(ctx, func(resp *cloudresourcemanager.ListProjectsResponse) error {
		for i := range resp.Projects {
			project := resp.Projects[i]
			if project.LifecycleState != "ACTIVE" {
				continue
			}
			projects = append(projects, project)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func listChildrenProjectsCall(projectSvc *cloudresourcemanager.ProjectsService, parent string) *cloudresourcemanager.ProjectsListCall {
	return projectSvc.List().Filter(parentFilter(parent))
}

func parentFilter(parent string) string {
	path := strings.Split(parent, "/")
	return fmt.Sprintf(`parent.type:%s parent.id:%s`, strings.TrimSuffix(path[0], "s"), path[1])
}

func listChildrenProjectsForEachParents(ctx context.Context, parentNames []string) ([]*cloudresourcemanager.Project, error) {
	var projects []*cloudresourcemanager.Project
	for _, parentName := range parentNames {
		pr, err := listChildrenProjects(ctx, parentName)
		if err != nil {
			return nil, err
		}
		projects = append(projects, pr...)
	}
	return projects, nil
}

func getBillingInfo(billingSvc *cloudbilling.ProjectsService, project *cloudresourcemanager.Project) (*cloudbilling.ProjectBillingInfo, error) {
	return billingSvc.GetBillingInfo("projects/" + project.ProjectId).Do()
}

func forEachProjectBillingInfo(ctx context.Context, projects []*cloudresourcemanager.Project,
	f func(*cloudresourcemanager.Project, *cloudbilling.ProjectBillingInfo)) error {
	svc, err := cloudbilling.NewService(ctx)
	if err != nil {
		return err
	}

	billingSvc := cloudbilling.NewProjectsService(svc)
	for _, project := range projects {
		billingInfo, err := getBillingInfo(billingSvc, project)
		if err != nil {
			return err
		}
		f(project, billingInfo)
	}
	return nil
}
