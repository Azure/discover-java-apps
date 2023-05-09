![](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)
![Coverage](https://github.com/Azure/discover-java-apps/blob/badge/badge.svg?branch=badge)

## What's this project for?

A script to discover java apps from your linux system by following steps:

1. SSH login to your linux system
2. Find the java process for java application
3. Collect information of runtime env, configuration, jar/war/ear files
4. Print the info to console (or specified file) in json or csv format

## Download and run

Download the binary files from [releases](https://github.com/Azure/azure-discovery-java-apps/releases)

- For Linux:

```bash
discovery-l -server 'servername' -port 'port' -username 'userwithsudo' -password 'password'
```

- For Windows:

```bash
discovery.exe -server 'servername' -port 'port' -username 'userwithsudo' -password 'password'
```

- For Mac (Intel chip):

```bash
discovery-darwin-amd64 -server 'servername' -port 'port' -username 'userwithsudo' -password 'password'
```

- For Mac (Apple silicon):

```bash
discovery-darwin-arm64 -server 'servername' -port 'port' -username 'userwithsudo' -password 'password'
```

> You can find the running log from __discovery.log__ in the same folder

## Sample output

The default output will be a json like

```javascript
[
  {
    "server": "127.0.0.1",
    // Application Name
    "appName": "hellospring",
    // Runnable artifact name
    "artifactName": "hellospring",
    //Spring Boot Version
    "springBootVersion": "1.5.14.RELEASE",
    //Application Type, for now only SpringBootExecutableFatJar supported, refer the definition from https://docs.spring.io/spring-boot/docs/current/reference/html/executable-jar.html
    "appType": "SpringBootFatJar",
    // Runtime JDK Version
    "runtimeJdkVersion": "17.0.6",
    // OS Name
    "OsName": "ubuntu",
    // OS Version
    "OsVersion": "2204",
    // Build JDK version
    "buildJdkVersion": "1.7",
    // Jar file location
    "jarFileLocation": "/home/user/hellospring1x-0.0.1-SNAPSHOT.jar",
    // Runtime Memory
    "jvmMemoryInMB": 128,
    // Application Port
    "appPort": 8080,
    "lastModifiedTime": "2023-02-05T09:24:40Z",
  },
  {
    ...
  }
]
```

CSV format is also supported when `-format csv` is received in command arguments
```csv
Server,AppName,AppType,AppPort,MavenArtifactGroup,MavenArtifact,MavenArtifactVersion,SpringBootVersion,BuildJdkVersion,RuntimeJdkVersion,HeapMemory(MB),OsName,OsVersion,JarFileLocation,JarFileSize(MB),JarFileModifiedTime
127.0.0.1,hellospring,SpringBoot,8080,com.example,hellospring,0.0.1-SNAPSHOT,2.4.13,8,10,128,ubuntu,18.04,/home/migrateadmin/hellospring2x-0.0.1-SNAPSHOT.jar,52,2022-11-21T08:08:46Z

```

## Contributing

We appreciate your help on the java app discovery. Before your contributing, please be noted:

1. Ensure you have Golang `1.20+` installed before starting try from source code
2. Run Makefile in `wsl` if you're Windows user
3. `70%` test coverage is mandatory in the PR build, so when you do PR, remember to include test cases as well.
4. Recommend to use [Ginkgo](https://onsi.github.io/ginkgo/) for a BDD style test case

## Build

```bash
make build
```

## Test

```bash
make test
```

## Check code coverage

```bash
go tool cover -func=coverage.out | grep total: | grep -Eo '[0-9]+\.[0-9]+'
```

## Limitation

Only support to discover the spring apps from Linux VM

## Road map

- More java app runtime are coming.

| Type           | Readiness | Ready Date |
|----------------| -- | -- |
| SpringBoot App | Ready | 2023-04 |
| Tomcat App     | Planned | - |
| WebLogic App   | Planned | - |
| WebSphere App  | Planned | - |
| JBoss EAP App  | Planned | - |

- More source operating systems are coming.

## Support

Report the issue to <https://github.com/Azure/azure-discovery-java-apps/issues>
