[![Docker Pulls](https://img.shields.io/docker/pulls/hellstromitltd/do-dyndns)](https://hub.docker.com/r/hellstromitltd/do-dyndns)

# do-dynds
Small Go application to automatically update a digitalocean DNS entry with your public IP (dyndns).

This application is intended to run as a docker container or a kubernetes deployment but can be ran standalone. It's also possible to deploy the container alongside your application by adding it to the same pod.

Containers can be found at: https://hub.docker.com/r/hellstromitltd/do-dyndns

# Why?
Exposing applications to the internet in an environment where you don't have a static IP can be tricky. DNS is usually an issue since your public IP can change at any time. There's multiple dyndns providers available but this application specifically handles a situation where your DNS is in DigitalOcean.

# Configuration
Check the `config.yml.example` file for an example configuration.

Create a digitalocean token at https://cloud.digitalocean.com/account/api/tokens and make sure it has **Write** permission. Optionally set an expiration, I'd suggest doing it but keep in mind that when the expiry date is reached you'll need to create a new token and update your configuration.

| Variable    | Description |
| ----------- | ----------- |
| interval    | Defines at what intervals we should check our public IP |
| ifconfig    | Holds the host and uri for the service used to check the public IP |
| digitalocean | Holds the digitalocean API token. The token can also be set as an environment variable `DO_TOKEN` |
| domains     | A list of domains that should be updated in DigitalOcean DNS |

In the example https://api.ipify.org/?format=json is used to query for the public IP but any service that gives a json response with the ip parameter should work.

## Preparing the DNS
Any missing A records will be automatically created by the application. Just make sure that the domain is setup in https://cloud.digitalocean.com/networking/domains

The application will ONLY update or create records that are defined in the config.yml. Other records that are manually added will be left alone.

# Deploying

## Compiling
It's possible to run the application standalone by compiling it manually (there's currently no pre-built binaries)

To compile the application checkout this repository:

```
git checkout https://github.com/HellstromIT/do-dyndns.git
```

Enter the app directory and run the following command
```
cd app
go build -o do-dyndns cmd/do-dyndns/main.go 
```

This should give you a binary called do-dyndns that you can copy to your path.

Note that since the application is currently focused on running within a docker container the config.yml file will need to be located in the `/data/` directory. That is the `data` directory in the root of your filesystem.


## Running the container
To run the container in docker there's two ways. The first is to run it with `docker run` and the second is using a `docker-compose` file.

In both cases you will need to mount the config.yml into the container at `/data/config.yml`.

### Docker run

Start by creating a config.yml file in your current directory. Use the included example as a base. Omit the digitalocean configuration if you plan to supply the token as an environment variable.

Then run the following command (token in config.yml)

```
docker run --name do-dyndns -d -v $(pwd)/config.yml:/data/config.yml hellstromitltd/do-dyndns:v0.1.3
```

Or this command if you prefer supplying the digitalocean token as an environment variable

```
docker run --name do-dyndns -d -v $(pwd)/config.yml:/data/config.yml --env DO_TOKEN=<your digitalocean token> hellstromitltd/do-dyndns:v0.1.3
```

The docker logs for the container will inform you if a domain has changed.

### Compose
A docker compose file is available in the repository and can be used to deploy the application. The config.yml is mounted into the container as a volume. By default the config.yml is expected to be located in the same directory as the docker-compose.yml but can be changed to another location if needed.

The DigitalOcean token can be specified in the config.yml or as an environment variable. If it's supplied in the config.yml just remove the environment part from the docker-comopose file.

To deploy copy the docker-compose.yml file from this repository (or clone the repository) and create a config.yml. Then run

```
docker-compose up -d
```

### Helm chart
A helm chart for deploying the service within Kubernetes will be provided in the future.
