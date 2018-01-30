## Zeppelin API client plugin for CI/CD workflow (Developer Guide)

### Build new docker image

    make docker

### Use example .env file and fill required vars

    cp .env.example .env

```
export PLUGIN_ZEPPELIN_ENDPOINT=""
export PLUGIN_ZEPPELIN_USERNAME=""
export PLUGIN_ZEPPELIN_PASSWORD=""
export PLUGIN_ZEPPELIN_NOTEBOOK_NAME=""
export PLUGIN_ZEPPELIN_NOTEBOOK_FILE_PATH=""
export PLUGIN_ZEPPELIN_NOTEBOOK_STATE="created"
```    

### Test container/plugin with docker

    docker run -env-file .env --rm -it banzaicloud/zeppelin_client:latest
