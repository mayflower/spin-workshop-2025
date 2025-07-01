# What is this?

This is the example project for our Workshop
"Running Web Assembly on the Cloud with Fermyon Spin".

The example application is an image repository. Images are uploaded to
the repository and can then converted to different size and formats
on-the-fly.

Please see [OVERVIEW.md](OVERVIEW.md) for more details.

# Dev container

We have prepared a Docker image `mayflower/spin-workshop-2025` that provides a
a ready-to-use development environment. You can start the container with

```
    $ docker run -it -p 3000:3000 mayflower/spin-workshop-2025
```

# Running

Once the container is up and running you can check out and run the application
within the container

```
    $ git clone https://github.com/mayflower/spin-workshop-2025
    $ cd spin-workshop-2025/image-repo
    $ spin up --build --listen 0.0.0.0:3000
```

The application is now listening on port 3000. Check [API.md](API.md) for the API of the
various components.

# Postman

There is a Postman collection and environment at the top level of the repo that you
can use to explore the API.

# Development

Once you have the dev container up and running you can use the VSCode
[Dev containers extension](https://code.visualstudio.com/docs/devcontainers/containers)
to attach to the running container and develop within the container.

# Exercise

The object of our workshop is redeveloping one of the application components
with Spin in the language of your choice. Delete your preferred component,
remove it from `spin.toml` and then get busy recreating and rewriting it from
scratch 😏 We're here to help you if you get stuck.

# Developing locally

If you want to pass on Docker you can develop and run the exercise locally on
Mac and Linux. On your own 😛 To do so you'll need the following prerequisites
(of course, you can skip on some by rewriting the respective component in
another language):

* [Rust + Cargo](https://www.rust-lang.org/tools/install) (best installed via `rustup`)
* The WASM/WASI target installed for rust: `rustup target add wasm32-wasip1`
* [Tinygo](https://tinygo.org/getting-started/install/) 0.18.0
* [NodeJS](https://nodejs.org/en/download) 22
* Python + `requirements.txt` installed via PIP (go for a venv if you value your
  Python installation)
* [Spin](https://spinframework.dev/v2/install)
* The [CRON trigger plugin](https://www.fermyon.com/blog/spin-cron-trigger) for Spin