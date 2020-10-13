# container-instance-metadata-server

The `container-instance-metadata-server` emulates the Cloud Run [container instance metadata server](https://cloud.google.com/run/docs/reference/container-contract#metadata-server) for a given [service account](https://cloud.google.com/iam/docs/service-accounts) and user supplied metadata.

Service accounts are [impersonated](https://cloud.google.com/iam/docs/understanding-service-accounts) using the [application default credentials](https://cloud.google.com/iam/docs/service-accounts#application_default_credentials) set by the `gcloud` commandline tool.

```
gcloud auth application-default login
```

## Available metadata server information

```
/computeMetadata/v1/instance/id
/computeMetadata/v1/instance/service-accounts/default/aliases
/computeMetadata/v1/instance/service-accounts/default/email
/computeMetadata/v1/instance/service-accounts/default/token
/computeMetadata/v1/instance/service-accounts/default/identity
/computeMetadata/v1/instance/region
/computeMetadata/v1/instance/zone
/computeMetadata/v1/project/numeric-project-id
/computeMetadata/v1/project/project-id
```

## Usage

```
container-instance-metadata-server -h
```
```
Usage of container-instance-metadata-server:
  -listen-address string
        The HTTP listen address (default "127.0.0.1:8888")
  -metadata string
        Metadata file path (default "metadata.json")
  -service-account string
        The email address of an IAM service account
```

## Tutorial

Create a metadata configuration file:

```
PROJECT_ID=$(gcloud config get-value project)
```

```
PROJECT_NUMBER=$(gcloud projects describe ${PROJECT_ID} --format="value(projectNumber)")
```

```
cat <<EOF > metadata.json
{
  "instance": {
    "region": "us-west1"
  },
  "project": {
    "numeric_project_id": "${PROJECT_NUMBER}",
    "project_id": "${PROJECT_ID}"
  }
}
EOF
```

Create the IAM service account that will be impersonated when generating access and identity tokens:

```
gcloud iam service-accounts create metadata-server
```

Create an IAM policy to grant the current logged in user the ability to impersonate the `metadata-server` service account:

```
cat <<EOF > policy.yaml
bindings:
  - role: "roles/iam.serviceAccountTokenCreator"
    members:
      - "user:$(gcloud config get-value account)"
EOF
```

Apply the IAM policy:

```
gcloud -q iam service-accounts set-iam-policy \
  "metadata-server@${PROJECT_ID}.iam.gserviceaccount.com" \
  policy.yaml
```

At this point you have the ability to impersonate the `metadata-server` service account.

> It may take a few minutes for the changes to take effect.

Start the `container-instance-metadata-server`:

```
container-instance-metadata-server \
  --metadata metadata.json \
  --service-account "metadata-server@${PROJECT_ID}.iam.gserviceaccount.com"
```

```
2020/10/13 00:16:07 Starting Container Instance Metadata Service ...
2020/10/13 00:16:07 Impersonating metadata-server@hightowerlabs.iam.gserviceaccount.com
2020/10/13 00:16:07 Listening on 127.0.0.1:8888
```

### Test

Retrieve the `instance/region` metadata key:

```
curl -i http://127.0.0.1:8888/computeMetadata/v1/instance/region \
  -H "Metadata-Flavor: Google"
```

```
HTTP/1.1 200 OK
Content-Type: application/text
Metadata-Flavor: Google
Server: Metadata Server for Serverless
Date: Mon, 12 Oct 2020 09:35:28 GMT
Content-Length: 8

us-west1
```

Generate an access token:

```
curl -i -G http://127.0.0.1:8888/computeMetadata/v1/instance/service-accounts/default/token \
  -H "Metadata-Flavor: Google" \
  --data-urlencode 'scopes=https://www.googleapis.com/auth/cloud-platform'
```

```
HTTP/1.1 200 OK
Content-Type: application/json
Metadata-Flavor: Google
Server: Metadata Server for Serverless
Date: Tue, 13 Oct 2020 08:51:25 GMT
Content-Length: 453

{"access_token":"redacted","expires_in":3599,"token_type":"Bearer"}
```

## Configuration

* `instance.region` string
* `project.numeric_project_id` string
* `project.project_id` string

### Example

```json
{
  "instance": {
    "region": "us-west1"
  },
  "project": {
    "numeric_project_id": "330612842442",
    "project_id": "hightowerlabs"
  }
}
```

## Routing Traffic

You can also run the metadata server on the same address `169.254.169.254` as Cloud Run does and also map the `metadata.google.internal` domain to that address.

Add a secondary IP address:

```
sudo ip address add 169.254.169.254/24 dev eth0
```

Append the following line to `/etc/hosts`:

```
169.254.169.254 metadata.google.internal
```

Resolve `metadata.google.internal`:

```
getent hosts metadata.google.internal
```

```
169.254.169.254 metadata.google.internal
```

Start the `container-instance-metadata-server`:

```
sudo container-instance-metadata-server \
  --listen-address "169.254.169.254:80" \
  --metadata metadata.json \
  --service-account "metadata-server@${PROJECT_ID}.iam.gserviceaccount.com"
```

### Test

```
curl -i http://metadata.google.internal/computeMetadata/v1/instance/id \
  -H "Metadata-Flavor: Google"
```

```
HTTP/1.1 200 OK
Content-Type: application/text
Metadata-Flavor: Google
Server: Metadata Server for Serverless
Date: Tue, 13 Oct 2020 08:52:52 GMT
Content-Length: 128

e368815d7aa80123751793efb5c86401d81edc3205f14ae196732d13805adcc65b38d1ef882a877a36526a52437acf3bc03c7b3f3bd7029e08020615724d7b74
```

> The id is auto generated by `container-instance-metadata-server` at start up.
