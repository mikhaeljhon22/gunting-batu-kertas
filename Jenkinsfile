pipeline{
    agent any

tools {
    go 'golang'
}
    stages{
        stage("Build"){
            steps{
                echo "Building the application..."
                sh "go build -o myapp ."
            }
        }
        stage("Running Test"){
            steps{
                echo "Running the test..."
                sh "go test ./..."
            }
        }
        stage("Run"){
            steps{
                echo "Running the application..."
                sh "./myapp"
            }
        }
        stage("Deploy"){
            steps{
                echo "Deploying the application..."
                // Deployment command go here
            }
        }
    }
}