# Get the list of deployed applications
deployments = cmo.getAppDeployments()

for deployment in deployments:
    deploymentName = deployment.getName()
    deploymentPath = deployment.getAbsoluteSourcePath()
    # Print the application name and its absolute path
    print("-----------------------------------------")
    print("application_name is: " + deploymentName + "; absolute_path is: " + deploymentPath + ";")
    print("-----------------------------------------")
    # Disconnect from the WebLogic Server disconnect()
disconnect()
