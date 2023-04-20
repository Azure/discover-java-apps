
# Project
A script to discover java apps from your linux system
1. SSH login to the your linux system 
2. Find the java process for spring application 
3. Collect the spring apps runtime env and config 
4. Print the info to console in json or csv format

## Run it from your local machine
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
<p>You can find the running log from discovery.log in the same folder

## Prerequisites

__Ensure you have Golang 1.20+ installed before starting try from source code__
__Run make in wsl if you're Windows user__

## Build

```bash
make build
```

## Test
```bash
make test
```

## Output
The default output will be a json like 
```json
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
    "jarFileLocation": "/home/migrateadmin/hellospring1x-0.0.1-SNAPSHOT.jar",
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

## Pull Request
70% percent test coverage is mandatory requirement when PR  

## Limitation
Only support to discover the spring apps from Linux VM

## Roadmap
1. More application types
   - Tomcat Apps support
   - WebLogic Apps
   - WebSphere Apps
   - JBoss EAP Apps
2. More Source system

## Support
Report the issue to https://github.com/Azure/azure-discovery-java-apps/issues
