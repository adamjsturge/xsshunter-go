# Contributing to xsshunter-go

Thanks for looking to contribute.

If you have a massive feature change I recommend putting as an issue before starting anything! That way we can discuss it before you start working on it.

## Where to start?

Look at the current issues and see if there is anything you want to do. I'm trying to put everything their even if I plan to work on it.

## Dev Environment

Actually super simple, we have a .vscode with recommended extensions and settings to make sure the code passes linter and scans.
Copy the .env.copy to .env then chaning the values to make sure it works for you. I notice that everychange you have to run 
```bash
docker compose build && docker compose up
```

Luckily go is fast and it shouldn't take too long for it to build

## Your own fork

Adding `DOCKER_USERNAME` and `DOCKER_PASSWORD`([DOCKER TOKEN](https://hub.docker.com/settings/security)) to your forked secrets will allow the pipeline to automatically push to docker and you can have your own docker image.
