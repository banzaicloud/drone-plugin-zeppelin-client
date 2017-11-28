## Zeppelin API client plugin for Drone

Zeppelin Notebook REST API client plugin for Drone. A step in the Pipeline PaaS CI/CD component to create / update / delete / run Zeppelin notebooks.

### Example drone config

.drone.yml

    pipeline:
      zeppelin:
        image: banzaicloud/pipeline_zeppelin_client:latest

        endpoint: http://[your-host-name-or-ip]/zeppelin
        zeppelin_username: admin
        zeppelin_password: *****
        log_level: info
        notebook_name: "sf-police-incidents"
        notebook_file_path: "sf_police_incidents.note.json"
        notebook_state: "running"

## Test container/plugin with drone exec

    drone exec --repo-name hello-world --workspace-path drone-test .drone.yml

## Build new docker image

    make docker

## For dev env push .env file

.env

    PLUGIN_ENDPOINT=http://[your-host-name-or-ip]/zeppelin
    PLUGIN_ZEPPELIN_USERNAME=admin
    PLUGIN_ZEPPELIN_PASSWORD=***
    PLUGIN_NOTEBOOK_NAME="example-notebook"
    PLUGIN_NOTEBOOK_FILE_PATH="example.note.json"
    PLUGIN_NOTEBOOK_STATE="created/running/deleted"
    PLUGIN_LOG_LEVEL=debug

### Test with `go run`

    go run -ldflags "-X main.version=1.0" main.go plugin.go log.go --plugin.log.level debug --plugin.log.format text