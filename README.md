# project42
Answer to the Ultimate Question of Life, the Universe, and Everything - Extension Project to the Cloud Foundry platform

## About project42
Cloud Foundry's most significant achievement is the 'cf push'-experience which simplifies the developers life in so many ways. But at day two, this experience gets lost because it gets well hidden in some CI/CD pipeline. Writing CI/CD pipelines can be hard and cumbersome.

Wouldn't it be nice to have the same experience of 'cf push' with CI/CD? But is that possible? But every CI/CD pipeline is different. 

With buildpacks we solved the same problem for application runtimes to get 'cf push'. So why do not do the same thing with CI/CD and fully integrate it in Cloud Foundry?

Benjamin Gandon and Christian Brinker present their work on project 42, their Cloud Foundry North America 2019 Summit's hackathon winning approach, until now.

project42 aims to extend the 'cf push' experience to CI/CD by creating standards in CI/CD through pipeline templates.

## Usage
To use project 42 you should have Cloud Foundry, an application already pushed to a Git repository accessible from a Concourse Server. For usage run the following steps:
1. Checkout the repository from Github. 
2. Install an configure Golang on your machine.
3. Install Cloud Foundry CLI matching your OS System
4. Run `go build && cf install-plugin project42 -f` inside of the repository folder
5. Run `cf careless-delivery --help` or `cf spin-up-prod` for further instructions
6. Run `cf careless-delivery` from your application's folder (with the application being pushed to a git repository with remote `origin/master`; actual we use this within Concourse and the CLI)

