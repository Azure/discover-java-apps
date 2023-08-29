#!/bin/bash

JAR_FILE=hellospring2x-0.0.1-SNAPSHOT.jar
sudo apt update
sudo apt install openjdk-11-jre-headless

if [ ! -f "$JAR_FILE" ]; then
    echo "$JAR_FILE does not exist, download from remote"
    curl -o $JAR_FILE https://raw.githubusercontent.com/Azure/discover-java-apps/main/mock/hellospring2x-0.0.1-SNAPSHOT.jar
fi
JAVA_OPTS_2x="-XX:InitialRAMPercentage=40.0 -XX:MaxRAMPercentage=40.0 -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=1099 -Dcom.sun.management.jmxremote.local.only=true -Dmanagement.endpoints.jmx.exposure.include=healthmetrics -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -DtestOption=abc=def -Dspring.jmx.enabled=true -Dserver.tomcat.mbeanregistry.enabled=true -Dfile.encoding=UTF8 -Dspring.config.import=optional:configserver:/ -Dspring.datasource.password=testpassword1234"
nohup java $JAVA_OPTS_2x -jar $JAR_FILE > app.log 2>&1 &