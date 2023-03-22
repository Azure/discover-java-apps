# Project
A script to discovery java apps in your linux system by serveral steps
1) login to the your linux system
2) find the java process for spring application
3) collect the spring apps runtime env and config
4) print the info to console

## Run it directly
Download the binary files from https://github.com/Azure/azure-discovery-java-apps/releases
1) For linux: execute cmd
- discovery-l -server 'servername' -port 'port' -username 'userwithsudo' -password 'password'
2) For windows: execute cmd 
- discovery.exe -server 'servername' -port 'port' -username 'userwithsudo' -password 'password'
<p>You can find the running log from discovery.log in the same folder

## Build the project
1) For linux, run build.sh
2) For Window, run build.bat
<p> Find your binanry file from bin folder

You can find the running log from discovery.log in the same folder

## Output
The default output will be a json like 
```json
[
  {
    // Application Name
    "appName": "hellospring",
    // Runnable artifact name
    "artifactName": "hellospring",
   // Checksum
    "checksum": "e66dea79cab7caf2ac0546ad6239d90252066bac19095acd06db149bc65fc3ae",
    //Spring Boot Version
    "springBootVersion": "1.5.14.RELEASE",
    //Application Type, for now only SpringBootExecutableFatJar supported, refer the definition from https://docs.spring.io/spring-boot/docs/current/reference/html/executable-jar.html
    "appType": "SpringBootFatJar",
    // Runtime JDK Version
    "runtimeJdkVersion": "17.0.6",
    // OS Version
    "OsVersion": "Linux version 5.15.0-1031-azure (buildd@lcy02-amd64-010) (gcc (Ubuntu 9.4.0-1ubuntu1~20.04.1) 9.4.0, GNU ld (GNU Binutils for Ubuntu) 2.34) #38~20.04.1-Ubuntu SMP Mon Jan 9 18:23:48 UTC 2023",
    // Build JDK version
    "buildJdkVersion": "1.7",
    //Runtime Env
    "environments": [
      "DBUS_SESSION_BUS_ADDRESS",
      "LOGNAME",
      "HOME"     
    ],
    // JVM Options
    "jvmOptions": [
      "-Xms128m",
      "-Dcom.sun.management.jmxremote.port",
      "-Dcom.sun.management.jmxremote.authenticate",
      "-Dcom.sun.management.jmxremote.ssl",
      "-DtestOption",
      "-Dspring.jmx.enabled",
      "-Dfile.encoding",
      "-Xmx128m",
      "-Dserver.tomcat.mbeanregistry.enabled",
      "-Dspring.config.import",
      "-Dcom.sun.management.jmxremote.local.only",
      "-Dmanagement.endpoints.jmx.exposure.include",
      "--server.port",
      "-Dcom.sun.management.jmxremote"
    ],
   // Dependencies of the application
    "dependencies": [
      "spring-web-4.3.18.RELEASE.jar",
      "spring-aop-4.3.18.RELEASE.jar",
      "spring-beans-4.3.18.RELEASE.jar",
      "spring-context-4.3.18.RELEASE.jar",
      "spring-webmvc-4.3.18.RELEASE.jar",
      "spring-expression-4.3.18.RELEASE.jar",
      "slf4j-api-1.7.25.jar",
      "spring-core-4.3.18.RELEASE.jar"
    ],
    // Jar file location
    "jarFileLocation": "/home/migrateadmin/hellospring1x-0.0.1-SNAPSHOT.jar",
    // Runtime Memory
    "jvmMemoryInMB": 128,
    // Application Port
    "appPort": 8080,
    // Other port
    "bindingPorts": [
      1099,
      8080,
      36385,
      37595
    ],
    "miscs": [
      {
        "key": "CONSOLE_OUTPUT_LOGGING_FILES",
        "value": ""
      },
      {
        "key": "LOGGING_FILES",
        "value": ""
      }
    ],
    "instanceCount": 1,
    "lastModifiedTime": "2023-02-05T09:24:40Z",
    "lastUpdatedTime": "2023-03-21T04:31:29Z"
  },
  {
    ...
  }
]
```

## Limitation
1) Only probe the spring apps from linux VM

## Roadmap
1) Tomcat Apps support
2) More target system

## Support
Report the issue to https://github.com/Azure/azure-discovery-java-apps/issues
