# User-manager

The user manager is the component in charge of managing users and roles in the system.

## Getting Started

Any information required by the final user to use this repo.

### Prerequisites

Detail any component that has to be installed to run this component.

* system-model: component responsible for maintaining user and role entities
* authx: component responsible for authentication 

### Build and compile

In order to build and compile this repository use the provided Makefile:

```
make all
```

This operation generates the binaries for this repo, download dependencies,
run existing tests and generate ready-to-deploy Kubernetes files.

### Run tests

Tests are executed using Ginkgo. To run all the available tests:

```
make test
```

### Update dependencies

Dependencies are managed using Godep. For an automatic dependencies download use:

```
make dep
```

In order to have all dependencies up-to-date run:

```
dep ensure -update -v
```

## User client interface
Explain the main features for the user client interface. Explaining the whole
CLI is never required. If you consider relevant to explain certain aspects of
this client, please provided the users with them.

Ignore this entry if it does not apply.

## Known Issues

Explain any relevant issues that may affect this repo.


## Contributing

Please read [contributing.md](contributing.md) for details on our code of conduct, and the process for submitting pull requests to us.


## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/nalej/user-manager/tags). 

## Authors

See also the list of [contributors](https://github.com/nalej/user-manager/contributors) who participated in this project.

## License
This project is licensed under the Apache 2.0 License - see the [LICENSE-2.0.txt](LICENSE-2.0.txt) file for details.


# user-manager

The user manager is the component in charge of managing users and roles in the system.

# Integration tests

The following table contains the variables that activate the integration tests

| Variable  | Example Value | Description |
| ------------- | ------------- |------------- |
| RUN_INTEGRATION_TEST  | true | Run integration tests |
| IT_SM_ADDRESS  | localhost:8800 | System Model Address |
| IT_AUTHX_ADDRESS  | localhost:8810 | Authx Address |