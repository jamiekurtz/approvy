This service, written with Google's Golang, is meant to facilitate remote approval for any kind of action. Approval is obtained via something like SMS, and returns approval status (i.e. approved or denied) to the caller.

If you're new to Golang, here are some resources:

- [Official Go website](https://golang.org/)
- [The Go Programming Language](https://www.amazon.com/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440)


# Overview

The general idea is that a caller, typically a shell script on some server somewhere, will request approval for a given action. In this case, we want the call to wait for the request to be approved (or, rejected). For example, a deployment script might want to request approval from a manager or deployment lead before continuing. In this way, an organization can configure just about any deployment tool, yet still be confident that if a deployment is initiated, it won't continue unless approval is obtained.

It is not uncommon for deployments to be fully automated these days. But that presents a problem to organizations that still need both approval and an audit trail of those approvals for deployments. And the person or group approving the deployment may not necessarily be an active user in the chosen deployment tool. As such, being able to approve or reject an approval request via SMS message is considerably more useful and efficient.

We can also use Approvy to help prevent accidental deployments. It is certainly not unheard of for an engineer or automation process to mistakenly deploy some release to the wrong target. Or maybe you want to provide your users an out-of-band approval mechanism for the big red DELETE button on their account. Any action or activity that would benefit from an SMS-based approval is a good candidate for Approvy.

If the deployment (or similar) script is leveraging the Approvy REST API via `curl` command, care must be taken to halt the deployment process if the approval request status is not `approved`. Or, you can checkout the sample script(s) in this repo for help with some wrapper code.


# Contributing

## Prerequisites

It is best to start by installing and configuring Go, then use the Go tools to clone this repo.

1. Install Go according to the instructions [here](https://golang.org/doc/install)
1. Make sure your GO workspace exists at $HOME/go

Once Go is installed and configured, run the following commands to clone the repo and download dependencies:

```
cd $HOME/go
go get github.com/jamiekurtz/approvy
```

Then you need to create your own config files:

```
cp config/config.yml.example config/config.yml
cp config/secrets.yml.example config/secrets.yml
```

Then to compile and run the app: 

```
cd $HOME/go/src/github.com/jamiekurtz/approvy
go run *.go
```

Browse to http://localhost:3000/status to make sure the site works.


## Other Things

With the service running, you can submit an approval request with the following CURL command:

```
curl -d 'from=bob&to=jamie&message=release 45 to production' http://localhost:3000/requests
```


# The API

All API requests must include an appropriate `X_API_KEY` HTTP header value.

### POST /requests

Submits a request for approval.

- from: name of requester
- to: name of approver
- message: text to appear in text message
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
- message: message of the approval request
- createdAt: date/time the request was submitted
- status: {waiting|expired|approved|rejected}
- responses: list of responses to this request
- completedAt: date/time the request is completed

### POST /requests/{id}/responses

Posts a response to a given approval request. 

- id: unique identifier of the approval request
- approved: {true|false}


