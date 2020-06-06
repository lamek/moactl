# MOA Command Line Tool

This project contains the `moactl` command line tool that simplifies the use of
the _MOA_.

## Getting started with `moactl`

> `moactl` is currently in private preview. Contact ... to request access.

### Prerequisites

Complete the following prerequisites.

#### Install the `aws-cli`

* Install the aws-cli https://aws.amazon.com/cli/
* Configure aws-cli to use the account that you would like to deploy the Amazon Red Hat OpenShift cluster into.

```
$ cat ~/.aws/credentials
[default]
aws_access_key_id = ...
aws_secret_access_key = ...

$ cat ~/.aws/config
[default]
output = table
region = us-east-2
```

Test your ability to query the AWS api:

```
$ aws ec2 describe-regions
```

#### Install `moactl`

Download and a release of the `moactl` command line utility and add it to your path.

> `moactl is under active developement. We recommend using the latest release.

```
curl https://github.com/openshift/moactl/releases/...
tar -cvzf moactl.tar.gz
export PATH=$PATH:$(pwd)
```

Add bash completion.

```
sudo ./moactl completion > /etc/bash_completion.d/moactl
. /etc/bash_completion.d/moactl

$ moactl
Command line tool for MOA.

Usage:
  moactl [command]

Available Commands:
  completion  Generates bash completion scripts
  create      Create a resource from stdin
  delete      Delete a specific resource
  describe    Show details of a specific resource
  edit        Edit a specific resource
  help        Help about any command
  init        Applies templates to support Managed OpenShift on AWS clusters
  list        List all resources of a specific type
  login       Log in to your Red Hat account
  logout      Log out
  logs        Show logs of a specific resource
  verify      Verify resources are configured correctly for cluster install
  version     Prints the version of the tool

Flags:
      --debug     Enable debug mode.
  -h, --help      help for moactl
  -v, --v Level   log level for V logs

Use "moactl [command] --help" for more information about a command.
```

#### Verify your AWS account permissions

Verify that your AWS account has the necessary permissions
*Link to MOA permissions docs*

```
$ moactl verify permissions
I: Validating SCP policies...
I: AWS SCP policies ok
```

Verify that your AWS account has necessary quota to deploy an OpenShift cluster.

> Sometimes quota varies by region, which may prevent you from deploying

```
$ export AWS_DEFAULT_REGION=us-west-2 && moactl verify quota
I: Validating AWS quota...
E: Insufficient AWS quotas
E: Service ebs quota code L-FD252861 Provisioned IOPS SSD (io1) volume storage not valid
```

If you receive and error, try another region:

```
$ export AWS_DEFAULT_REGION=us-east-2 && moactl verify quota
I: Validating AWS quota...
I: AWS quota ok
```

If both the permissions and quota checks pass, proceed to initializing your AWS account.

#### Initialize your AWS account

The following step runs a CloudFormation template that prepares your AWS account for OpenShift deployment and management.  It typically takes 1-2 minutes to complete.

```
$ moactl init
I: Logged in as 'rh-moa-user' on 'https://api.openshift.com'
I: Validating AWS credentials...
I: AWS credentials are valid!
I: Validating SCP policies...
I: AWS SCP policies ok
I: Validating AWS quota...
I: AWS quota ok
I: Ensuring cluster administrator user 'osdCcsAdmin'...
I: Admin user 'osdCcsAdmin' created successfully!
I: Verifying whether OpenShift command-line tool is available...
E: OpenShift command-line tool is not installed.
Go to https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/ to download the OpenShift client and add it to your PATH.
```

#### Install the `oc` client

Install [the latest OpenShift command line utility](https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/), also known as `oc`.

Verify `oc` is installed correctly.

```
$ oc version --client
Client Version: 4.4.0-202005231254-4a4cd75
```

### Create a cluster

```
$ time moactl create cluster --name=rh-moa-test1
I: Creating cluster with identifier '1de87g7c30g75qechgh7l5b2bha6r04e' and name 'rh-moa-test1'
I: To view list of clusters and their status, run `moactl list clusters`
I: Cluster 'rh-moa-test1' has been created.
I: Once the cluster is 'Ready' you will need to add an Identity Provider and define the list of cluster administrators. See `moactl create idp --help` and `moactl create user --help` for more information.
I: To determine when your cluster is Ready, run `moactl describe cluster rh-moa-test1`.
```

The cluster creation process can take up to 40 minutes, during which the State will transition from pending to installing, and finally to ready.

```
$ moactl describe cluster rh-moa-test1
Name:        rh-moa-test1
ID:          1de87g7c30g75qechgh7l5b2bha6r04e
External ID: 34322be7-b2a7-45c2-af39-2c684ce624e1
API URL:     https://api.rh-moa-test1.j9n4.s1.devshift.org:6443
Console URL: https://console-openshift-console.apps.rh-moa-test1.j9n4.s1.devshift.org
Nodes:       Master: 3, Infra: 3, Compute: 4
Region:      us-east-2
State:       ready
Created:     May 27, 2020
```

If installation fails or does not change to ready after 40 minutes, proceed to the Installation debugging document.

### Access your cluster

In order to login to your cluster, you must configure an Identity Provider and then create a user. This guide covers creating a `dedicated-admin` user and then creating a `cluster-admin` user.

#### Configure an IDP

This example uses GitHub as and Identity Provider.

> To view other supported Identity Providers run `moactl create idp --help`

Run the following command to create an Identiy Provider backed by GitHub. Follow the instructions in the output to register your application with GitHub.

```bash
$ moactl create idp --cluster rh-moa-test1 --type github
I: Loading cluster 'rh-moa-test1'
I: Loading identity providers for cluster 'rh-moa-test1'
To use GitHub as an identity provider, you must first register the application:
? List of GitHub organizations or teams that will have access to this cluster: openshift-online
* Open the following URL: https://github.com/organizations/openshift-online/settings/applications/new?oauth_application%5Bcallback_url%5D=https%3A%2F%2Foauth-openshift.apps.rh-moa-test1.j9n4.s
1.devshift.org%2Foauth2callback%2Fgithub-1&oauth_application%5Bname%5D=rh-moa-test1&oauth_application%5Burl%5D=https%3A%2F%2Fconsole-openshift-console.apps.rh-moa-test1.j9n4.s1.devshift.org
* Click on 'Register application'
? Copy the Client ID provided by GitHub: <client-id-from-GitHub>
? Copy the Client Secret provided by GitHub: <client-secret-from-GitHub>
I: Configuring IDP for cluster 'rh-moa-test1'
I: Identity Provider 'github-1' has been created. You need to ensure that there is a list of cluster administrators defined. See `moactl user add --help` for more information. To login into th
e console, open https://console-openshift-console.apps.rh-moa-test1.j9n4.s1.devshift.org and click on github-1
```

> The IDP can take 1-2 minutes to be configured within your cluster.

Check to see that your identity provider has been configured.

```
$ moactl list idps --cluster rh-moa-test1
NAME        TYPE      AUTH URL
github-1    GitHub    https://oauth-openshift.apps.rh-moa-test1.j9n4.s1.devshift.org/oauth2callback/github-1
```

#### Create a `dedicated-admin` user

Now promote your github user to Dedicated Admin *link to dedicated admin docs*

$ moactl create user --cluster rh-moa-test1 --dedicated-admins=jeremyeder

At this point you should be able to login to your cluster using your github ID.

Get the console URL of your cluster:

```
$ moactl describe cluster rh-moa-test1
Name:        rh-moa-test1
ID:          1de87g7c30g75qechgh7l5b2bha6r04e
External ID: 34322be7-b2a7-45c2-af39-2c684ce624e1
API URL:     https://api.rh-moa-test1.j9n4.s1.devshift.org:6443
Console URL: https://console-openshift-console.apps.rh-moa-test1.j9n4.s1.devshift.org
Nodes:       Master: 3, Infra: 3, Compute: 4
Region:      us-east-2
State:       ready
Created:     May 27, 2020
```

Login to the cluster using oc

In the top right of the OpenShift console, click your name and click Copy Login Command.  Now click github-1 and finally click Display Token.
Copy and paste the oc login command into your terminal.

```
$ oc login --token=z3sgOGVDk0k4vbqo_wFqBQQTnT-nA-nQLb8XEmWnw4X --server=https://api.rh-moa-test1.j9n4.s1.devshift.org:6443
Logged into "https://api.rh-moa-test1.j9n4.s1.devshift.org:6443" as "jeremyeder" using the token provided.

You have access to 67 projects, the list has been suppressed. You can list all projects with 'oc projects'

Using project "default".
```

Run a simple command to verify everything is setup properly and you are logged in:

```
$ oc version
Client Version: 4.4.0-202005231254-4a4cd75
Server Version: 4.3.18
Kubernetes Version: v1.16.2
```

Now try to get something I shouldn't be able to as dedicated-admin, and indeed some of it fails as expected.

```
$ oc get all -n openshift-apiserver
NAME                  READY   STATUS    RESTARTS   AGE
pod/apiserver-6ndg2   1/1     Running   0          17h
pod/apiserver-lrmxs   1/1     Running   0          17h
pod/apiserver-tsqhz   1/1     Running   0          17h
NAME          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/api   ClusterIP   172.30.23.241   <none>        443/TCP   17h
NAME                       DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR                     AGE
daemonset.apps/apiserver   3         3         3       3            3           node-role.kubernetes.io/master=   17h
Error from server (Forbidden): horizontalpodautoscalers.autoscaling is forbidden: User "jeremyeder" cannot list resource "horizontalpodautoscalers" in API group "autoscaling" in the namespace 
"openshift-apiserver"
Error from server (Forbidden): jobs.batch is forbidden: User "jeremyeder" cannot list resource "jobs" in API group "batch" in the namespace "openshift-apiserver"
Error from server (Forbidden): cronjobs.batch is forbidden: User "jeremyeder" cannot list resource "cronjobs" in API group "batch" in the namespace "openshift-apiserver"
Error from server (Forbidden): imagestreams.image.openshift.io is forbidden: User "jeremyeder" cannot list resource "imagestreams" in API group "image.openshift.io" in the namespace "openshift
-apiserver"
```
#### Enable `cluster-admin` and create a `cluster-admin` user

First enable `cluster-admin` capability on the cluster:

```bash
$ moactl edit cluster rh-moa-test1 --enable-cluster-admins
```

Now give myself cluster-admin

```
$ moactl create user --cluster rh-moa-test1 --cluster-admins jeremyeder
$ moactl list users --cluster rh-moa-test1
GROUP             NAME
cluster-admins    jeremyeder
dedicated-admins  jeremyeder
```

Now retry the previous command that failed as dedicated-admin:

```
$ oc get all -n openshift-apiserver                       
NAME                  READY   STATUS    RESTARTS   AGE
pod/apiserver-6ndg2   1/1     Running   0          17h
pod/apiserver-lrmxs   1/1     Running   0          17h
pod/apiserver-tsqhz   1/1     Running   0          17h
NAME          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/api   ClusterIP   172.30.23.241   <none>        443/TCP   18h
NAME                       DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR                     AGE
daemonset.apps/apiserver   3         3         3       3            3           node-role.kubernetes.io/master=   18h
```

### Deleting your cluster

When you are done with your cluster you can delete it.

```bash
$ moactl delete cluster -c rh-moa-test1
```

To delete the CloudFormation as well, run the following command:

```bash
$ moactl init --delete-stack
```