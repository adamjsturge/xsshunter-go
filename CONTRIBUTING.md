# Contributing to xsshunter-go

Thanks for looking to contribute.

If you have a massive feature change I recommend putting as an issue before starting anything! That way we can discuss it before you start working on it.

## Where to start?

Look at the current issues and see if there is anything you want to do. I'm trying to put everything their even if I plan to work on it.

## Dev Environment

Start by forking the repo!

Actually super simple, we have a .vscode with recommended extensions and settings to make sure the code passes linter and scans.
Copy the .env.copy to .env then chaning the values to make sure it works for you. I notice that everychange you have to run 
```bash
git clone git@github.com:yourgithubusernamefork/xsshunter-go.git
cd xsshunter-go
cp .env.copy .env
docker compose up --build
```

Luckily go is fast and it shouldn't take too long for it to build

## Git Workflow

Please make a new branch for every feature or bug fix you are working on. This makes it easier to review and merge your code. We also merge into branch `dev` and then `main` is merged from `dev` when we are ready to release. So please make sure your PR is against `dev`.

## Code Style

I'm using the golangci-lint to make sure the code is formatted correctly. If you are using vscode you can install the extension and it will format the code for you. If you are not using vscode you can run `golangci-lint run` to make sure the code is formatted correctly.

## Commit Messages

Please make sure your commit messages are clear and concise. If you are fixing a bug please include the issue number in the commit message. If you are adding a new feature please include a brief description of the feature.

## Pull Requests

Please make sure your PR is against the `dev` branch. Please make sure your PR is clear and concise. If you are fixing a bug please include the issue number in the PR. If you are adding a new feature please include a brief description of the feature.

## Testing

I'm trying to add tests to everything I can. If you are adding a new feature please add tests to it. If you are fixing a bug please add a test to make sure it doesn't happen again. All tests are in the e2e folder and are run with playwright.

## Your own fork

Adding `DOCKER_USERNAME` and `DOCKER_PASSWORD`([DOCKER TOKEN](https://hub.docker.com/settings/security)) to your forked secrets will allow the pipeline to automatically push to docker and you can have your own docker image.
