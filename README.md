This service, written with Google's Golang, is meant to facilitate remote approval for any kind of action. Approval is obtained via something like SMS, and returns approval status (i.e. approved or denied) to the caller.

If you're new to Golang, here are some resources:

- [Official Go website](https://golang.org/)
- [The Go Programming Language](https://www.amazon.com/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440)


# Overview

The general idea is that a caller, typically a shell script on some server somewhere, will request approval for a given action. In this case, we want the call to wait for the request to be approved (or, rejected). For example, a deployment script might want to request approval from a manager or deployment lead before continuing. In this way, an organization can configure just about any deployment tool, yet still be confident that if a deployment is initiated, it won't continue unless approval is obtained.

It is not uncommon for deployments to be fully automated these days. But that presents a problem to organizations that still need both approval and an audit trail of those approvals for deployments. And the person or group approving the deployment may not necessarily be an active user in the chosen deployment tool. As such, being able to approve or reject an approval request via SMS message is considerably more useful and efficient.

If the deployment (or similar) script is leveraging the Approvy REST API via `curl` command, care must be taken to halt the deployment process if the approval request status is not `approved`. Or, you can checkout the sample script(s) in this repo for help with some wrapper code.


# Contributing

## Prerequisites

Make sure you have the following installed:

- [Vagrant](https://www.vagrantup.com/)
- [VirtualBox](https://www.virtualbox.org/wiki/Downloads)

Using Vagrant to spin up a development VM in VirtualBox is the **only** supported configuration for working on this application.


## Preparing (or, resetting) the environment

Open your terminal to the repo's root, and enter the following:

    vagrant up

This command will download (if needed) and build an Ubuntu-based virtual machine. It will take maybe 5 or 10 minutes the first time.

Next, we need to install dependencies and create some sample/test data by running the following:

1. From your terminal where you ran `vagrant up`, run `vagrant ssh`
1. Once you're in the SSH session, change the current directory: `cd $APPROVYPATH`
1. Run the following to load all dev/test data: `setup/reset.sh`

At this point, all of the necessary dependencies should be installed and the test data should be loaded for you to develop and test against a local 
instance of the site. Read the next section for help on starting the site and associated worker processes.

To exit out of the VM simply hit `ctrl d`.

To stop the VM after you've exited the SSH session, enter the following: `vagrant halt`

To completely delete the VM, enter the following: `vagrant destroy`


## Compile and Run

Start by ensuring you're in the Vagrant box with: `vagrant ssh`. Then ensure you're in the correct directory on the box with: `cd $APPROVYPATH`. And also, you need to have already run `setup/reset.sh`, per the previous section.

To install dependencies and compile the app: `go install`

That command will place the binary (i.e. `approvy`) into your `$GOPATH/bin` directory.

Then, to run the server: `approvy`

Then use your browser to navigate to [http://localhost:3000](http://localhost:3000).


## Other Things

In addition to the Redis server, the Redis tools are also installed in the Vagrant VM. For example, you can access the Redis client with `redis-cli`.

With the service running, you can submit an approval request with the following CURL command:

```
curl -d 'from=bob&to=jamie&subject=release 45 to production' http://localhost:3000/requests
```


# The API

All API requests must include an appropriate `X_API_KEY` HTTP header value.

### POST /requests

Submits a request for approval.

- from: name of requester
- to: name of approver
- subject: text to appear in text message
- waitOnRedirect: indicates you want the redirect to include the `wait=yes` parameter
- expirationDurationSeconds: number of seconds until the request is considered expired

Responds with 301 to details of the resulting request. The redirected location will be a GET to /requests/{id}?wait={yes|no}

### GET /requests/{id}?wait=yes

- id: unique ID of the appoval request (as returned by the POST to /requests)
- wait: 'yes' indicates you want the call to block until the request is completed (i.e. approved or rejected)

Response:

- id: unique identifier of the approval request
- from: name of the sender
- to: name of the approver (person or group)
- subject: subject of the approval request
- createdAt: date/time the request was submitted
- status: {waiting|expired|approved|rejected}
- responses: list of reponses to this request
- completedAt: date/time the request is completed

### POST /requests/{id}/responses

Posts a response to a given approval request. 

- id: unique identifier of the approval request
- response: {approve|reject}


