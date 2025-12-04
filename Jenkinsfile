pipeline {
    agent any

    tools {
        go 'golang'
    }

    stages {
        stage("Build") {
            steps {
                echo "Building the application..."
                sh "go build -o myapp ."
            }
        }

        stage("Running Test") {
            steps {
                echo "Running the test..."
                sh "go test ./..."
            }
        }

        stage("Run") {
            steps {
                echo "Running the application..."

                // Jalankan myapp di background + simpan PID
                sh '''
                    ./myapp &
                    echo $! > app.pid
                    echo "Application started with PID $(cat app.pid)"

                    # Tunggu 5 detik (opsional – untuk smoke test)
                    sleep 5

                    echo "Stopping application..."
                    kill $(cat app.pid)
                '''
            }
        }

        stage("Deploy") {
            steps {
                echo "Deploying the application..."
                // Deployment commands here
            }
        }
    }
}
