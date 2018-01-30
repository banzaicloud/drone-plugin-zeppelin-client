
## Zeppelin API client plugin for CI/CD workflow

Zeppelin REST API client plugin for CI/CD workflow. A step in the Pipeline PaaS CI/CD component to create / update / delete and also run a Zeppelin notebook via Zeppelin's REST API.
You had to specify desired state of Zeppelin notebook in zeppelin_notebook_state option.
These are the valid states:

- present - notebook will created or updated
- absent - notebook will be deleted if exists
- running - notebook will be created or updated and started

#### Specify required secrets

Provide valid credentials for the pipeline API.

These options needs to be specified in the CI/CD [GUI](https://github.com/banzaicloud/pipeline/blob/master/docs/pipeline-howto.md#cicd-secrets).

* plugin_zeppelin_username: Zeppelin username
* plugin_zeppelin_password: Zeppelin password

### Main options

| Option                       | Description                                    | Default  | Required |
| -------------                | -----------------------                        | --------:| --------:|
| zeppelin_notebook_name       | Name of the notebook to be created / updated   | ""       | Yes       |
| zeppelin_notebook_file_path  | Path to notebook file                          | ""       | Yes       |
| zeppelin_notebook_state      | Desired state of notebook                      | ""       | Yes       |
| zeppelin_endpoint            | Zeppelin REST API endpoint to use              | ""       | Yes       |

### Example YAML

```
run:
    image: banzaicloud/k8s-proxy:0.2.0
    original_image: banzaicloud/zeppelin-client:0.2.0
    zeppelin_notebook_name: "sf-police-incidents"
    zeppelin_notebook_file_path: "sf_police_incidents.note.json"
    zeppelin_notebook_state: "running"
    zeppelin_endpoint: "http://release-1-zeppelin:8080/zeppelin"

    secrets: [ plugin_zeppelin_username, plugin_zeppelin_password ]
```

For a full example of provisioning Zeppelin Server then run a notebook on it checkout [Zeppelin PDI example](https://github.com/banzaicloud/zeppelin-pdi-example).
Are you a developer? Click [here](dev.md)
