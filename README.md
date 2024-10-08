# UnlockEdv2

## Requirements

Currently, UnlockEdv2 is tested on Windows (WSL), Mac (homebrew) and Linux.

- Docker && Docker Compose
- Go 1.23
- Node.js > 18.0
- Yarn
- Make (optional, but recommended)

## Development

If you would like to contribute, please have a look at our [contribution guidelines](CONTRIBUTING.md).

### Dependencies (Local)

- Go 1.23
- Node.js > 18.0
- Docker && Docker-Compose

### Dependencies (Deployment/Production)

- Node.js > 18.0
- Postgres 16.0

- Clone the repository
- `cp .env.example .env && cp frontend/.env.example frontend/.env`
- run `make help` for complete instructions
- run `cd frontend && yarn prepare`

**For frontend development:**

- Run `make frontend-dev`

This will build everything but the client, which you can then run separately with `yarn run dev` in the frontend directory.

**For backend development:**

- Run `make backend-dev`

This will build only the Auth and run Postgres. You are responsible for building and running the server, middleware, and optionally the client/vite.

**For 'Production':**

- Run `make prod` to build the entire project in docker. You can then go to `localhost` in your browser.
  NOTE: This is not a true production build, but simply a way to run the entire project in containers

Login with `SuperAdmin` and password: `ChangeMe!`

You will be prompted immediately to set a new password and name for the default facility, and then you
will be redirected to the dashboard.

**Integrations:**

- _Kolibri_:
  If you wish to run + develop against `Kolibri` locally:
  first, you need to build and tag the docker image

  - `docker buildx build . -f config/kolibri.Dockerfile -t unlockedlabs.org/kolibri:latest`
  - Run `make kolibri`
    This will run all the containers, similarly to the 'prod' command but with the addition of the kolibri container,
    which will be available at `localhost:8000`

  The server will automatically create an OIDC client, which you can access either by logging in for the first
  time and going to `/api/oidc/clients` and adding the client_id and client_secret as environment variables in
  the `config/docker-compose.kolibri.yml` file. After you set this value, run `docker restart unlockedv2-kolibri-1`
  and you should be able to access the provider at `localhost:8000`

### To migrate the database to a fresh state, run `make migrate-fresh` (you can do this while docker is running with all the services, but you must restart the server (e.g. `docker restart unlockedv2-server-1` if your repo directory is called UnlockEdv2)

### To seed the database with some basic test data, run `make seed`

### **Quick fixes to common issues with development**

This product is under **active** development, and things are subject to abrupt change, frequently.
Here are a couple things to try if you are having errors with your local development environment:
If these don't work, look at the FAQ below, and if those don't work feel free to open an issue 👍

**First**, no matter the problem: try clearing your cookies in your browser. We rely on a csrf_token, an ory_session_token and an unlocked_token
to validate and authenticate our users. Clearing them and logging back in will often solve problems we have when developing
and creating many sessions and restarting the server.

**Second**: try migrating the database to a fresh state, and restarting docker.

**Third**: make sure your `.env` file is up to date. It's possible that you copied your .env.example over upon first cloning the repo,
so be sure that it didn't update in a later commit.

## Style/Linting

- Naming and style convention is outlined in our CONTRIBUTING.md file.

#### Proper linting/formatting _will_ run automatically in a git hook before each commit. If you want to run them beforehand, you can `cd` into frontend and run `npx prettier -w .` or `cd backend` and `gofmt -w .` IF for some reason you need to skip the pre-commit hooks, you can run `git commit --no-verify`but _do not_ do this unless you know what you are doing and you understand the CI/CD will fail if you submit a PR

# FAQ/Troubleshooting

### Why is docker not starting properly?

> Chances are, this is a permissions issue with docker. If you are new to docker, you may need to run `sudo usermod -aG docker $USER`
> to add yourself to the docker group. You will need to log out and back in for this to take effect.
> Try starting docker with `sudo dockerd`, or restarting the daemon with `sudo systemctl restart docker.service`, followed by `docker run hello-world`
> to ensure docker is running properly before again trying any of the `make` or `docker compose` commands.

### Docker says "network {UUID} not found"

run `docker compose up {-f docker-compose.yml} {-f config/docker-compose.fe-dev.yml | -f config/docker-compose.prod.yml} --build --force-recreate` depending on what services you need

# Debugging

Two tools you can use to aid in debugging:

- React Developer Tools. This is available for several browser flavors and available on their website: <https://react.dev/learn/react-developer-tools>

# License

UnlockEdv2 is open-sourced software licensed under the [Apache License, Version 2.0](https://opensource.org/license/apache-2-0/).
