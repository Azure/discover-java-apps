package mock

import (
	"fmt"
)

var (
	ExecutableProcess   = fmt.Sprintf(" %d 1000 java -javaagent:/path/to/applicationinsights.jar -XX:InitialRAMPercentage=60.0 -XX:MaxRAMPercentage=60.0 -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=1099 -Dcom.sun.management.jmxremote.password=testpassword1234 -Dcom.sun.management.jmxremote.local.only=true -Dmanagement.endpoints.jmx.exposure.include=health,metrics -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -DtestOption=abc=def -Dspring.jmx.enabled=true -Dserver.tomcat.mbeanregistry.enabled=true -Dfile.encoding=UTF8 -Dspring.config.import=optional:configserver:/ -Dcom.sun.management.jmxremote.password=testpassword1234 -jar %s", ExecutableProcessId, ExecutableJarFile)
	SpringBoot2xProcess = fmt.Sprintf("%d     0 /usr/bin/qemu-x86_64 /usr/bin/java -javaagent:/path/to/applicationinsights.jar -Xmx128m -Xms128m -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=1099 -Dcom.sun.management.jmxremote.local.only=true -Dmanagement.endpoints.jmx.exposure.include=healthmetrics -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -DtestOption=abc=def -Dspring.jmx.enabled=true -Dserver.tomcat.mbeanregistry.enabled=true -Dfile.encoding=UTF8 -Dspring.config.import=optional:configserver:/ -Dcom.sun.management.jmxremote.password=testpassword1234 -jar %s", SpringBoot2xProcessId, SpringBoot2xJarFileLocation)
	SpringBoot1xProcess = fmt.Sprintf("%d     0 /usr/bin/qemu-x86_64 /usr/bin/java -javaagent:/path/to/applicationinsights.jar -Xmx128m -Xms128m -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=1099 -Dcom.sun.management.jmxremote.local.only=true -Dmanagement.endpoints.jmx.exposure.include=healthmetrics -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -DtestOption=abc=def -Dspring.jmx.enabled=true -Dserver.tomcat.mbeanregistry.enabled=true -Dfile.encoding=UTF8 -Dspring.config.import=optional:configserver:/ -Dcom.sun.management.jmxremote.password=testpassword1234 -jar %s", SpringBoot1xProcessId, SpringBoot1xJarFileLocation)
	ErrorProcess        = fmt.Sprintf("%d     0 /usr/bin/qemu-x86_64 /usr/bin/java -javaagent:/path/to/applicationinsights.jar -Xmx128m -Xms128m -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=1099 -Dcom.sun.management.jmxremote.local.only=true -Dmanagement.endpoints.jmx.exposure.include=healthmetrics -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -DtestOption=abc=def -Dspring.jmx.enabled=true -Dserver.tomcat.mbeanregistry.enabled=true -Dfile.encoding=UTF8 -Dspring.config.import=optional:configserver:/ -Dcom.sun.management.jmxremote.password=testpassword1234 -jar %s", ErrorProcessId, SpringBoot1xJarFileLocation)

	Manifest = `Manifest-Version: 1.0
Created-By: Maven Jar Plugin 3.2.0
Build-Jdk-Spec: 11
Implementation-Title: hellospring
Implementation-Version: 0.0.1-SNAPSHOT
Main-Class: org.springframework.boot.loader.JarLauncher
Start-Class: com.example.hellospring.HellospringApplication
Spring-Boot-Version: 2.4.13
Spring-Boot-Classes: BOOT-INF/classes/
Spring-Boot-Lib: BOOT-INF/lib/
Spring-Boot-Classpath-Index: BOOT-INF/classpath.idx
Spring-Boot-Layers-Index: BOOT-INF/layers.idx
`

	TestEnv = "SHELL=/bin/bash\000PWD=/home/azureuser\000LOGNAME=azureuser\000XDG_SESSION_TYPE=tty\000HOME=/home/azureuser\000LANG=C.UTF-8\000LS_COLORS=rs=0:di=01;34:ln=01;36:mh=00:pi=40;33:so=01;35:do=01;35:bd=40;33;01:cd=40;33;01:or=40;31;01:mi=00:su=37;41:sg=30;43:ca=30;41:tw=30;42:ow=34;42:st=37;44:ex=01;32:*.tar=01;31:*.tgz=01;31:*.arc=01;31:*.arj=01;31:*.taz=01;31:*.lha=01;31:*.lz4=01;31:*.lzh=01;31:*.lzma=01;31:*.tlz=01;31:*.txz=01;31:*.tzo=01;31:*.t7z=01;31:*.zip=01;31:*.z=01;31:*.dz=01;31:*.gz=01;31:*.lrz=01;31:*.lz=01;31:*.lzo=01;31:*.xz=01;31:*.zst=01;31:*.tzst=01;31:*.bz2=01;31:*.bz=01;31:*.tbz=01;31:*.tbz2=01;31:*.tz=01;31:*.deb=01;31:*.rpm=01;31:*.jar=01;31:*.war=01;31:*.ear=01;31:*.sar=01;31:*.rar=01;31:*.alz=01;31:*.ace=01;31:*.zoo=01;31:*.cpio=01;31:*.7z=01;31:*.rz=01;31:*.cab=01;31:*.wim=01;31:*.swm=01;31:*.dwm=01;31:*.esd=01;31:*.jpg=01;35:*.jpeg=01;35:*.mjpg=01;35:*.mjpeg=01;35:*.gif=01;35:*.bmp=01;35:*.pbm=01;35:*.pgm=01;35:*.ppm=01;35:*.tga=01;35:*.xbm=01;35:*.xpm=01;35:*.tif=01;35:*.tiff=01;35:*.png=01;35:*.svg=01;35:*.svgz=01;35:*.mng=01;35:*.pcx=01;35:*.mov=01;35:*.mpg=01;35:*.mpeg=01;35:*.m2v=01;35:*.mkv=01;35:*.webm=01;35:*.ogm=01;35:*.mp4=01;35:*.m4v=01;35:*.mp4v=01;35:*.vob=01;35:*.qt=01;35:*.nuv=01;35:*.wmv=01;35:*.asf=01;35:*.rm=01;35:*.rmvb=01;35:*.flc=01;35:*.avi=01;35:*.fli=01;35:*.flv=01;35:*.gl=01;35:*.dl=01;35:*.xcf=01;35:*.xwd=01;35:*.yuv=01;35:*.cgm=01;35:*.emf=01;35:*.ogv=01;35:*.ogx=01;35:*.aac=00;36:*.au=00;36:*.flac=00;36:*.m4a=00;36:*.mid=00;36:*.midi=00;36:*.mka=00;36:*.mp3=00;36:*.mpc=00;36:*.ogg=00;36:*.ra=00;36:*.wav=00;36:*.oga=00;36:*.opus=00;36:*.spx=00;36:*.xspf=00;36:\000SSH_CONNECTION=20.210.124.35 57406 10.5.0.4 22\000XDG_SESSION_CLASS=user\000TERM=xterm-256color\000USER=azureuser\000SHLVL=1\000XDG_SESSION_ID=184\000XDG_RUNTIME_DIR=/run/user/1000\000SSH_CLIENT=20.210.124.35 57406 22\000PATH=/usr/local/bin:/usr/bin:/bin:/usr/local/games:/usr/games\000MAIL=/var/mail/azureuser\000SSH_TTY=/dev/pts/0\000_=/usr/bin/env\000_=/usr/bin/env\000test_option=test\000DB_PASSWORD=testpassword1234"

	TestJvmOptions = []string{
		"-XX:InitialRAMPercentage=60.0",
		"-XX:MaxRAMPercentage=60.0",
		"-Dcom.sun.management.jmxremote",
		"-Dcom.sun.management.jmxremote.password=testpassword1234",
		"-Dcom.sun.management.jmxremote.port=1099",
		"-Dcom.sun.management.jmxremote.local.only=true",
		"-Dmanagement.endpoints.jmx.exposure.include=health,metrics",
		"-Dcom.sun.management.jmxremote.authenticate=false",
		"-Dcom.sun.management.jmxremote.ssl=false",
		"-DtestOption=abc=def",
		"-Dspring.jmx.enabled=true",
		"-Dserver.tomcat.mbeanregistry.enabled=true",
		"-Dfile.encoding=UTF8",
		"-Dspring.application.name=hellospring2x",
		"-Dspring.config.import=optional:configserver:/",
		"-jar",
	}

	TotalMemory = "  987654321\n"

	DefaultMaxHeapSize = "  987654321\n"

	Checksum = "  987654321\n"

	RuntimeJdkVersion = " 11.0.16_232\n"

	SpringBoot1xAppName = "hellospring1x"
	SpringBoot2xAppName = "hellospring2x"
	ExecutableAppName   = "executable"
	SpringBoot1xCrName  = "hellospring1x-hellospring"
	SpringBoot2xCrName  = "hellospring2x-hellospring"
	ExecutableCrName    = "executable-executable"
	SpringBoot1xJarFile = "hellospring1x-0.0.1-SNAPSHOT.jar"
	SpringBoot2xJarFile = "hellospring2x-0.0.1-SNAPSHOT.jar"
	ExecutableJarFile   = "executable-0.0.1-SNAPSHOT.jar"
	SpringBoot1xVersion = "1.5.14.RELEASE"
	SpringBoot2xVersion = "2.4.13"
	Jdk7Version         = "1.7"
	Jdk8Version         = "8"

	SpringBoot1xProcessId = 1024
	SpringBoot2xProcessId = 1
	ExecutableProcessId   = 5647
	ErrorProcessId        = 999

	SpringBoot2xJarFileLocation = fmt.Sprintf("/home/azure/%s", SpringBoot2xJarFile)
	SpringBoot1xJarFileLocation = fmt.Sprintf("/home/azure/%s", SpringBoot1xJarFile)
	ExecutableJarFileLocation   = fmt.Sprintf("/home/azure/%s", ExecutableJarFile)

	Ports = "  22\n 8080\n38193\n44981"

	Host = "centos-8-openjdk11"

	Pom = `
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
	<modelVersion>4.0.0</modelVersion>
	<parent>
		<groupId>org.springframework.boot</groupId>
		<artifactId>spring-boot-starter-parent</artifactId>
		<version>2.4.13</version>
		<relativePath/> <!-- lookup parent from repository -->
	</parent>
	<groupId>com.example</groupId>
	<artifactId>hellospring2x</artifactId>
	<version>0.0.1-SNAPSHOT</version>
	<name>hellospring</name>
	<description>Demo project for Spring Boot</description>
	<properties>
		<java.version>8</java.version>
		<spring-cloud.version>2020.0.3</spring-cloud.version>
	</properties>

	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-actuator</artifactId>
			<exclusions>
				<exclusion>
					<groupId>org.junit.vintage</groupId>
					<artifactId>junit-vintage-engine</artifactId>
				</exclusion>
			</exclusions>
		</dependency>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
		</dependency>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-data-redis</artifactId>
		</dependency>
		<!-- https://mvnrepository.com/artifact/org.springframework.boot/spring-boot-starter-quartz -->
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-quartz</artifactId>
		</dependency>
		<dependency>
			<groupId>org.springframework.cloud</groupId>
			<artifactId>spring-cloud-sleuth-zipkin</artifactId>
		</dependency>
		<dependency>
			<groupId>org.springframework.cloud</groupId>
			<artifactId>spring-cloud-starter-config</artifactId>
		</dependency>
		<dependency>
			<groupId>org.springframework.cloud</groupId>
			<artifactId>spring-cloud-starter-netflix-eureka-client</artifactId>
			<exclusions>
				<exclusion>
					<groupId>junit</groupId>
					<artifactId>junit</artifactId>
				</exclusion>
			</exclusions>
		</dependency>
		<dependency>
			<groupId>org.springframework.cloud</groupId>
			<artifactId>spring-cloud-starter-sleuth</artifactId>
		</dependency>
	</dependencies>
	<dependencyManagement>
		<dependencies>
			<dependency>
				<groupId>org.springframework.cloud</groupId>
				<artifactId>spring-cloud-dependencies</artifactId>
				<version>${spring-cloud.version}</version>
				<type>pom</type>
				<scope>import</scope>
			</dependency>
		</dependencies>
	</dependencyManagement>
</project>
`

	ApplicationYaml = `spring:
  application:
    name: hellospring
    port: 8080
    value: 11.22
    password: ${SPRING_APPLICATION_PASSWORD}
  datasource:
    password: testpassword1234
    url: jdbc:h2:dev
    username: SA

config:
  list:
    - name: password
      value: testpass1234

---
spring:
  config:
    activate:
      on-profile: staging
  datasource:
    credential: 'password'
    url: jdbc:h2:staging
    username: SA
`
	ApplicationProperties = `
spring.application.name=hellospring1x
spring.cloud.config.import-check.enabled=false
spring.cloud.config.enabled=false

# jdbc
spring.datasource.hikari.jdbc-url=jdbc:mysql://localhost:3306/test?useUnicode=true&characterEncoding=utf8&useSSL=false
spring.datasource.password=testpassword1234
`
)
