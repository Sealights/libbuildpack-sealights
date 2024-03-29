# libbuildpack-sealights
Cloud Foundry Buildpack integrations with Sealights

## Bind your application to your Sealights service

1. First step is to create and configure user provided service

    For Linux:
    ```cf cups sealights -p '{"token":"ey…"}'```
    For Windows:
    ```cf cups sealights -p "{\"token\":\"ey…\"}"```

    Note: you can change prameters later with command `cf uups sealights -p ...`

    Complete list of the prameters currently supported by the buildpack service is:
    ```
    {
        "version"               // sealights version. default value: 'latest'
        "verb"                  // allow to specify command for the agent. default value: 'startBackgroundTestListener'
        "customAgentUrl"        // sealights agent will be downloaded from this url if provided
        "customCommand"         // allow to replace application start command
        "proxy"                 // proxy for the agent download client
        "proxyUsername"         // proxy user
        "proxyPassword"         // proxy password

        + rest of the parameters will be passed directly to the Sealights agent
    }
    ```

2. Bind your application to your Sealights service

    cf bind-service [app name] sealights

3. Restage an application to apply the changes

    cf restage [app name]

## Logs

You can enable Debug logs level by setting `BP_DEBUG` env variable:
```
cf set-env <your-app> BP_DEBUG True
```