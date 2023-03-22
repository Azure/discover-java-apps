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

## Run the project
1) For linux: execute cmd output/discovery-l -server 'servername' -port 'port' -username 'userwithsudo' -password 'password'
2) For windows: execute cmd output/discovery.exe -server 'servername' -port 'port' -username 'userwithsudo' -password 'password'

You can find the running log from discovery.log in the same folder

## Output
The default output will be a json like 
```json
[
  {
    "appName": "Name of App",
    "artifactName": "artifact name of the file",
    "checksum": "checksum number of the file，sample e66dea79cab7caf2ac0546ad6239d90252066bac19095acd06db149bc65fc3ae",
    "springBootVersion": "spring boot version, sample 2.4.5",
    "appType": "Application Type like SpringBootExecutableFatJar",
    "runtimeJdkVersion": "runtime JDK version,sample 17.0.6",
    "OsVersion": "Linux version 5.15.0-1031-azure (buildd@lcy02-amd64-010) (gcc (Ubuntu 9.4.0-1ubuntu1~20.04.1) 9.4.0, GNU ld (GNU Binutils for Ubuntu) 2.34) #38~20.04.1-Ubuntu SMP Mon Jan 9 18:23:48 UTC 2023",
    "buildJdkVersion": "application build JDK sample 17",
    "environments": [
      'env key list'
    ],
    "jvmOptions": [
     // JVM options key list 
      "-Xms128m",
      "-Dcom.sun.management.jmxremote.port",
      "-Dcom.sun.management.jmxremote.authenticate",
      "-Dcom.sun.management.jmxremote.ssl"
    ],
    "dependencies": [
      // depdency list
      "spring-core-4.3.18.RELEASE.jar"
    ],
    "jarFileLocation": "java file path,sample /home/migrateadmin/hellospring1x-0.0.1-SNAPSHOT.jar",
    "jvmMemoryInMB": 128,
    "appPort": 8080,
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
    // instance number
    "instanceCount": 1,
    "lastModifiedTime": "2023-02-05T09:24:40Z",
    "lastUpdatedTime": "2023-03-21T04:31:29Z"
  },
  {
    "appName": "hellospring",
    "artifactName": "hellospring",
    "checksum": "a54c2663c1d6f6ec1c411bd68dd4eaca15d91ac198ddf21efd9790defb114c62",
    "springBootVersion": "2.4.13",
    "appType": "SpringBootFatJar",
    "runtimeJdkVersion": "17.0.6",
    "OsVersion": "Linux version 5.15.0-1031-azure (buildd@lcy02-amd64-010) (gcc (Ubuntu 9.4.0-1ubuntu1~20.04.1) 9.4.0, GNU ld (GNU Binutils for Ubuntu) 2.34) #38~20.04.1-Ubuntu SMP Mon Jan 9 18:23:48 UTC 2023",
    "buildJdkVersion": "8",
    "environments": [
      "XDG_SESSION_TYPE",
      "LANG",
      "LS_COLORS",
      "SGX_AESM_ADDR",
      "TERM",
      "XDG_DATA_DIRS",
      "LOGNAME",
      "HOME",
      "SHLVL",
      "PATH",
      "SHELL",
      "PWD",
      "SSH_CONNECTION",
      "XDG_SESSION_CLASS",
      "SSH_TTY",
      "_",
      "MOTD_SHOWN",
      "USER",
      "XDG_SESSION_ID",
      "XDG_RUNTIME_DIR",
      "SSH_CLIENT",
      "DBUS_SESSION_BUS_ADDRESS"
    ],
    "jvmOptions": [
      "-XX:InitialRAMPercentage",
      "-Dspring.jmx.enabled",
      "-Dspring.datasource.password",
      "-Dmanagement.endpoints.jmx.exposure.include",
      "-Dserver.tomcat.mbeanregistry.enabled",
      "-Dspring.config.import",
      "-Dcom.sun.management.jmxremote.authenticate",
      "-Dcom.sun.management.jmxremote.ssl",
      "-Dcom.sun.management.jmxremote.port",
      "-Dcom.sun.management.jmxremote.local.only",
      "-DtestOption",
      "-Dfile.encoding",
      "--server.port",
      "-XX:MaxRAMPercentage",
      "-Dcom.sun.management.jmxremote"
    ],
    "dependencies": [
      "spring-boot-2.4.13.jar",
      "spring-boot-autoconfigure-2.4.13.jar",
      "logback-classic-1.2.7.jar",
      "logback-core-1.2.7.jar",
      "log4j-to-slf4j-2.13.3.jar",
      "log4j-api-2.13.3.jar",
      "jul-to-slf4j-1.7.32.jar",
      "jakarta.annotation-api-1.3.5.jar",
      "spring-core-5.3.13.jar",
      "spring-jcl-5.3.13.jar",
      "snakeyaml-1.27.jar",
      "spring-boot-actuator-autoconfigure-2.4.13.jar",
      "spring-boot-actuator-2.4.13.jar",
      "jackson-datatype-jsr310-2.11.4.jar",
      "micrometer-core-1.6.13.jar",
      "HdrHistogram-2.1.12.jar",
      "LatencyUtils-2.0.3.jar",
      "jackson-datatype-jdk8-2.11.4.jar",
      "jackson-module-parameter-names-2.11.4.jar",
      "tomcat-embed-core-9.0.55.jar",
      "jakarta.el-3.0.4.jar",
      "tomcat-embed-websocket-9.0.55.jar",
      "spring-web-5.3.13.jar",
      "spring-beans-5.3.13.jar",
      "spring-webmvc-5.3.13.jar",
      "spring-aop-5.3.13.jar",
      "spring-context-5.3.13.jar",
      "spring-expression-5.3.13.jar",
      "spring-data-redis-2.4.15.jar",
      "spring-data-keyvalue-2.4.15.jar",
      "spring-data-commons-2.4.15.jar",
      "spring-oxm-5.3.13.jar",
      "slf4j-api-1.7.32.jar",
      "lettuce-core-6.0.8.RELEASE.jar",
      "netty-common-4.1.70.Final.jar",
      "netty-handler-4.1.70.Final.jar",
      "netty-resolver-4.1.70.Final.jar",
      "netty-buffer-4.1.70.Final.jar",
      "netty-codec-4.1.70.Final.jar",
      "netty-transport-4.1.70.Final.jar",
      "reactor-core-3.4.12.jar",
      "reactive-streams-1.0.3.jar",
      "spring-context-support-5.3.13.jar",
      "spring-tx-5.3.13.jar",
      "quartz-2.3.2.jar",
      "mchange-commons-java-0.2.15.jar",
      "spring-cloud-sleuth-zipkin-3.0.3.jar",
      "spring-cloud-sleuth-instrumentation-3.0.3.jar",
      "spring-cloud-sleuth-api-3.0.3.jar",
      "aspectjrt-1.9.7.jar",
      "spring-cloud-commons-3.0.3.jar",
      "spring-security-crypto-5.4.9.jar",
      "zipkin-2.23.0.jar",
      "zipkin-reporter-2.16.1.jar",
      "zipkin-reporter-brave-2.16.1.jar",
      "zipkin-sender-kafka-2.16.1.jar",
      "zipkin-sender-activemq-client-2.16.1.jar",
      "zipkin-sender-amqp-client-2.16.1.jar",
      "spring-cloud-starter-config-3.0.4.jar",
      "spring-cloud-starter-3.0.3.jar",
      "spring-cloud-context-3.0.3.jar",
      "spring-security-rsa-1.0.10.RELEASE.jar",
      "bcpkix-jdk15on-1.68.jar",
      "bcprov-jdk15on-1.68.jar",
      "spring-cloud-config-client-3.0.4.jar",
      "jackson-annotations-2.11.4.jar",
      "httpclient-4.5.13.jar",
      "httpcore-4.4.14.jar",
      "commons-codec-1.15.jar",
      "jackson-databind-2.11.4.jar",
      "jackson-core-2.11.4.jar",
      "spring-cloud-starter-netflix-eureka-client-3.0.3.jar",
      "spring-cloud-netflix-eureka-client-3.0.3.jar",
      "eureka-client-1.10.14.jar",
      "netflix-eventbus-0.3.0.jar",
      "netflix-infix-0.3.0.jar",
      "commons-jxpath-1.3.jar",
      "joda-time-2.3.jar",
      "antlr-runtime-3.4.jar",
      "stringtemplate-3.2.1.jar",
      "antlr-2.7.7.jar",
      "gson-2.8.9.jar",
      "commons-math-2.2.jar",
      "xstream-1.4.16.jar",
      "mxparser-1.2.1.jar",
      "xmlpull-1.1.3.1.jar",
      "jsr311-api-1.1.1.jar",
      "servo-core-0.12.21.jar",
      "guava-19.0.jar",
      "commons-configuration-1.10.jar",
      "commons-lang-2.6.jar",
      "guice-4.1.0.jar",
      "javax.inject-1.jar",
      "aopalliance-1.0.jar",
      "jettison-1.4.0.jar",
      "eureka-core-1.10.14.jar",
      "woodstox-core-6.2.1.jar",
      "stax2-api-4.2.1.jar",
      "spring-cloud-starter-loadbalancer-3.0.3.jar",
      "spring-cloud-loadbalancer-3.0.3.jar",
      "hibernate-validator-6.1.7.Final.jar",
      "jakarta.validation-api-2.0.2.jar",
      "jboss-logging-3.4.2.Final.jar",
      "classmate-1.5.1.jar",
      "reactor-extra-3.4.5.jar",
      "evictor-1.0.0.jar",
      "spring-cloud-starter-sleuth-3.0.3.jar",
      "aspectjweaver-1.9.7.jar",
      "spring-cloud-sleuth-autoconfigure-3.0.3.jar",
      "spring-cloud-sleuth-brave-3.0.3.jar",
      "brave-5.13.2.jar",
      "brave-context-slf4j-5.13.2.jar",
      "brave-instrumentation-messaging-5.13.2.jar",
      "brave-instrumentation-rpc-5.13.2.jar",
      "brave-instrumentation-spring-rabbit-5.13.2.jar",
      "brave-instrumentation-kafka-clients-5.13.2.jar",
      "brave-instrumentation-kafka-streams-5.13.2.jar",
      "brave-instrumentation-httpclient-5.13.2.jar",
      "brave-instrumentation-http-5.13.2.jar",
      "brave-instrumentation-httpasyncclient-5.13.2.jar",
      "brave-instrumentation-jms-5.13.2.jar",
      "brave-instrumentation-mongodb-5.13.2.jar",
      "brave-propagation-aws-0.21.3.jar",
      "zipkin-reporter-metrics-micrometer-2.16.1.jar",
      "spring-boot-jarmode-layertools-2.4.13.jar"
    ],
    "jarFileLocation": "/home/migrateadmin/hellospring2x-0.0.1-SNAPSHOT.jar",
    "jvmMemoryInMB": 9601,
    "appPort": 8081,
    "bindingPorts": [
      1199,
      8081,
      33293,
      42239
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
    "lastModifiedTime": "2023-02-05T09:25:19Z",
    "lastUpdatedTime": "2023-03-21T04:31:38Z"
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
