/*
Copyright (c) 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ocm

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	errors "github.com/zgalor/weberr"

	"github.com/openshift/moactl/pkg/ocm/properties"
)

// Regular expression to used to make sure that the identifier or name given by the user is
// safe and that it there is no risk of SQL injection:
var clusterKeyRE = regexp.MustCompile(`^(\w|-)+$`)

func IsValidClusterKey(clusterKey string) bool {
	return clusterKeyRE.MatchString(clusterKey)
}

func HasClusters(client *cmv1.ClustersClient, creatorARN string) (bool, error) {
	query := fmt.Sprintf("properties.%s = '%s'", properties.CreatorARN, creatorARN)
	response, err := client.List().
		Search(query).
		Page(1).
		Size(1).
		Send()
	if err != nil {
		return false, fmt.Errorf("Failed to list clusters: %v", err)
	}

	return response.Total() > 0, nil
}

func GetCluster(client *cmv1.ClustersClient, clusterKey string, creatorARN string) (*cmv1.Cluster, error) {
	query := fmt.Sprintf(
		"(id = '%s' or name = '%s') and properties.%s = '%s'",
		clusterKey, clusterKey, properties.CreatorARN, creatorARN,
	)
	response, err := client.List().
		Search(query).
		Page(1).
		Size(1).
		Send()
	if err != nil {
		return nil, fmt.Errorf("Failed to locate cluster '%s': %v", clusterKey, err)
	}

	switch response.Total() {
	case 0:
		return nil, fmt.Errorf("There is no cluster with identifier or name '%s'", clusterKey)
	case 1:
		return response.Items().Slice()[0], nil
	default:
		return nil, fmt.Errorf("There are %d clusters with identifier or name '%s'", response.Total(), clusterKey)
	}
}

func GetIdentityProviders(client *cmv1.ClustersClient, clusterID string) ([]*cmv1.IdentityProvider, error) {
	idpClient := client.Cluster(clusterID).IdentityProviders()
	response, err := idpClient.List().
		Page(1).
		Size(-1).
		Send()
	if err != nil {
		return nil, fmt.Errorf("Failed to get identity providers for cluster '%s': %v", clusterID, err)
	}

	return response.Items().Slice(), nil
}

func GetIngresses(client *cmv1.ClustersClient, clusterID string) ([]*cmv1.Ingress, error) {
	ingressClient := client.Cluster(clusterID).Ingresses()
	response, err := ingressClient.List().
		Page(1).
		Size(-1).
		Send()
	if err != nil {
		return nil, fmt.Errorf("Failed to get ingresses for cluster '%s': %v", clusterID, err)
	}

	return response.Items().Slice(), nil
}

func GetUsers(client *cmv1.ClustersClient, clusterID string, group string) ([]*cmv1.User, error) {
	usersClient := client.Cluster(clusterID).Groups().Group(group).Users()
	response, err := usersClient.List().
		Page(1).
		Size(-1).
		Send()
	if err != nil {
		return nil, fmt.Errorf("Failed to get %s users for cluster '%s': %v", group, clusterID, err)
	}

	return response.Items().Slice(), nil
}

func GetLogs(client *cmv1.ClustersClient, clusterID string, tail int) (logs *cmv1.Log, err error) {
	logsClient := client.Cluster(clusterID).Logs().Log("hive")
	response, err := logsClient.Get().
		Parameter("tail", tail).
		Send()
	if err != nil {
		err = fmt.Errorf("Failed to get logs for cluster '%s': %v", clusterID, err)
		if response.Status() == http.StatusNotFound {
			err = errors.NotFound.UserErrorf("Failed to get logs for cluster '%s'", clusterID)
		}
		return
	}

	return response.Body(), nil
}

func PollLogs(client *cmv1.ClustersClient, clusterID string,
	cb func(*cmv1.LogGetResponse) bool) (logs *cmv1.Log, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer func() {
		cancel()
	}()

	logsClient := client.Cluster(clusterID).Logs().Log("hive")
	response, err := logsClient.Poll().
		Parameter("tail", 100).
		Interval(5 * time.Second).
		Predicate(cb).
		StartContext(ctx)
	if err != nil {
		err = fmt.Errorf("Failed to poll logs for cluster '%s': %v", clusterID, err)
		if response.Status() == http.StatusNotFound {
			err = errors.NotFound.UserErrorf("Failed to poll logs for cluster '%s'", clusterID)
		}
		return
	}

	return response.Body(), nil
}
