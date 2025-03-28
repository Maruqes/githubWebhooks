
# githubWebhooks

GitHub Webhooks Auto-Updater is a Golang application that listens for GitHub webhooks and automatically updates specified local repositories. This README provides the necessary instructions to configure and use this tool.

## Requirements
- Go 1.18 or higher installed on your system.
- An environment capable of running HTTP applications.

## Configuration
1. Clone this repository to your local environment:
  

2. Create a `.env` file in the root of the project with the following format:
   ```env
   REPO_PATH0="/path/to/repo1"
   REPO_PATH1="/path/to/repo2"
   REPO_PATH2="/path/to/repo3"
   SECRET="your-webhook-secret"
   PORT="8080"
   ```
   - **REPO_PATHn**: Absolute paths to the repositories that should be automatically updated.
   - **SECRET**: The secret used to validate the webhooks received from GitHub.
   - **PORT**: The port on which the server will run.

3. Ensure that each specified repository is properly configured and accessible by the user running this application.


4. The server will listen on the port specified in the `.env` file. Ensure that your GitHub repository's webhooks are configured to send events to `http://<your-domain>:<port>/webhook`.

5. Configure pm2 processes **THE PROCESS NAMES SHOULD BE THE SAME AS THE REPOSITORY NAME**, this project restarts the process on pm2 based on the name
   
## Setting Up Webhooks on GitHub
1. On GitHub, go to the settings page of your repository.
2. Click **Webhooks** > **Add webhook**.
3. Fill in the fields with the following information:
   - **Payload URL**: `http://<your-domain>:<port>/webhook`
   - **Content type**: `application/x-www-form-urlencoded`
   - **Secret**: Use the value of the `SECRET` field from your `.env` file.
4. Select the events you want to send or choose `Send me everything`, only tested with `only the push`.
5. Click **Add webhook**.

## How It Works
- Whenever a specified event is sent to the configured endpoint, the application validates the webhook using the provided secret.
- If the event is valid, the application executes the necessary commands to update the configured repositories.

## Customization
- To add more repositories, include additional `REPO_PATHn` variables in the `.env` file.
- Adjust the code as needed to support specific workflows.

## License
This project is licensed under the GNU GENERAL PUBLIC LICENSE. See the `LICENSE` file for more details.
