![](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)
![Coverage](https://github.com/Azure/discover-java-apps/blob/badge/badge.svg?branch=badge)

## What's this project for?

A script to discover java apps from your linux system by following steps:

1. SSH login to your linux system
2. Find the process for weblogic  server
3. Collect information of runtime env, configuration, jar/war/ear files
4. Print the info to console (or specified file) in json or csv format

## Download and run

Download the binary files from [releases](https://github.com/Azure/discover-java-apps/releases)

- For Linux:

```bash
weblogic-l -server 'servername' -port 'port' -username 'userwithsudo' -password 'password' -weblogicusername "weblogic" -weblogicpassword "weblogicpassword" 
```

- For Windows:

```bash
weblogic.exe -server 'servername' -port 'port' -username 'userwithsudo' -password 'password' -weblogicusername "weblogic" -weblogicpassword "weblogicpassword"
```

- For Mac (Intel chip):

```bash
weblogic_darwin_amd64 -server 'servername' -port 'port' -username 'userwithsudo' -password 'password' -weblogicusername "weblogic" -weblogicpassword "weblogicpassword"
```

- For Mac (Apple silicon):

```bash
weblogic_darwin_arm64 -server 'servername' -port 'port' -username 'userwithsudo' -password 'password' -weblogicusername "weblogic" -weblogicpassword "weblogicpassword"
```

> You can find the running log from __weblogic.log__ in the same folder

## Sample output

The default output will be a json like

```javascript
[
  {
    "server": "20.39.48.129",
    // Application Name
    "appName": "shoppingcart",
    "appType": "war",
    // Application Port
    "appPort": 7001,
    // Runtime Memory
    "jvmMemoryInMB": 512,
    // OS Name
    "osName": "ol",
    // OS Version
    "osVersion": "7.6",
    // appliation absolute path
    "jarFileLocation": "/u01/domains/adminDomain/servers/admin/upload/shoppingcart.war/app/shoppingcart.war",
    "lastModifiedTime": "2023-06-01T06:04:21Z",
    // Weblogic Version
    "weblogicVersion": "14.1.1.0.0",
    "runtimeJdkVersion": "11.0.11",
    // Weblogic Patches
    "weblogicPatches": "32697788;24150631;Mon Jan 09 07:56:07 UTC 2023;WLS PATCH SET UPDATE 14.1.1.0.210329\n32581868;24146453;Mon Jan 09 07:55:16 UTC 2023;Bundle patch for Oracle Coherence Version 14.1.1.0.4\n",
    "deploymentTarget": "admin",
    "serverType": "weblogic"
  },
  {
    ...
  }
]
```

CSV format is also supported when `-format csv` is received in command arguments
```csv
Server,AppName,AppType,AppPort,JvmHeapMemory(MB),OsName,OsVersion,JarFileLocation,JarFileModifiedTime,WeblogicVersion,WeblogicPatches,DeploymentTarget,RuntimeJdkVersion,ServerType
20.39.48.129,testwebapp,war,7001,512,ol,7.6,/u01/domains/adminDomain/servers/admin/upload/testwebapp.war/app/testwebapp.war,2023-05-29T07:24:19Z,14.1.1.0.0,"32697788;24150631;Mon Jan 09 07:56:07 UTC 2023;WLS PATCH SET UPDATE 14.1.1.0.210329
32581868;24146453;Mon Jan 09 07:55:16 UTC 2023;Bundle patch for Oracle Coherence Version 14.1.1.0.4
",admin,11.0.11,weblogic
20.39.48.129,shoppingcart,war,7001,512,ol,7.6,/u01/domains/adminDomain/servers/admin/upload/shoppingcart.war/app/shoppingcart.war,2023-06-01T06:04:21Z,14.1.1.0.0,"32697788;24150631;Mon Jan 09 07:56:07 UTC 2023;WLS PATCH SET UPDATE 14.1.1.0.210329
32581868;24146453;Mon Jan 09 07:55:16 UTC 2023;Bundle patch for Oracle Coherence Version 14.1.1.0.4
",admin,11.0.11,weblogic


```

## Contributing

We appreciate your help on the java app weblogic. Before your contributing, please be noted:

1. Ensure you have Golang `1.20+` installed before starting try from source code
2. Run Makefile in `wsl` if you're Windows user
3. `70%` test coverage is mandatory in the PR build, so when you do PR, remember to include test cases as well.
4. Recommend to use [Ginkgo](https://onsi.github.io/ginkgo/) for a BDD style test case

## Build

```bash
make build
```

## Check code coverage

```bash
go tool cover -func=coverage.out | grep total: | grep -Eo '[0-9]+\.[0-9]+'
```

## Limitation

Only support to discover the weblogic apps from Linux VM

## Road map

- More java app runtime are coming.

| Type           | Readiness | Ready Date |
|----------------| -- | -- |
| SpringBoot App | Ready | 2023-04 |
| WebLogic App   | Alpha | 2023-06 |
| Tomcat App     | Planned | - |
| WebSphere App  | Planned | - |
| JBoss EAP App  | Planned | - |

- More source operating systems are coming.

## Support

Report the issue to <https://github.com/Azure/azure-weblogic-java-apps/issues>
