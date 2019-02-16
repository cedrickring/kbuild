# ![Logo](logo/kbuild.png)

Build container images inside a Kubernetes Cluster with your local build context.

For more information see: [Kaniko](https://github.com/GoogleContainerTools/kaniko)

### Requirements
1. `~/.docker/config.json` exists and authenticated to a registry
2. `~/.kube/config` exists and configured with the correct cluster
3. A Kubernetes Cluster
4. A Container Registry

### Installation

For Linux or Mac use

```bash
curl -fsSL https://raw.githubusercontent.com/cedrickring/kbuild/master/scripts/get | bash
```

### Usage

Build an image of the current directory with tag `repository:tag`

```bash
kbuild -t repository:tag
````

To specify a Dockerfile in the working directory, use:

```bash
kbuild -t repository:tag -d Dockerfile.dev
```

or

```bash
kbuild --tag repository:tag --dockerfile Dockerfile.dev
```

respectively.

You can specify multiple image tags by repeating the tag flag.

### Additional Flags
 
#### -w / --workdir

Specify the working directory (defaults to the directory you're currently in)

#### -d / --dockerfile

Path to the `Dockerfile` in the build context (defaults to `Dockerfile`)

#### -c / --cache

Enable `RUN` command caching for faster builds (See [here](https://github.com/GoogleContainerTools/kaniko/blob/master/README.md#--cache))

#### --cache-repo

Specify the repo to cache build steps in (defaults to `<repository>cache`, repo retrieved from the image tag)

#### -n / --namespace

Specify namespace for the builder to run in (defaults to "default" namespace)

#### --build-arg

This flag allows you to pass in build args (ARG) for the Kaniko executor

#### --bucket

The bucket to use for [Google Cloud Storage](#google-cloud-storage)

### Registry credentials

You can either have your Docker Container Registry credentials in your `~/.docker/config.json` or provide them with the
`--username` (`-u`) and `--password` (`-p`) flags.

When using the cli flags, the container registry url is guessed based on the first provided image tag.
e.g. `-t my.registry.com/tag` is guessed as `my.registry.com`. If no specific registry is provided in the tag, it defaults to
`https://index.docker.io/v1/`.

### Google Cloud Storage 

If you want to use the Google Cloud Storage to store your build context, you have to pass `gcs` as the first argument to kbuild
and specify the `--bucket` to use.

Example: `kbuild -t image:tag --bucket mybucket gcs`

You might need to create [a service account key](https://console.cloud.google.com/apis/credentials/serviceaccountkey) and store the path to the `service-account.json` in the `GOOGLE_APPLICATION_CREDENTIALS` environment variable. 

### How does kbuild work?

In order to use the local context, the context needs to be tar-ed, copied to an Init Container, which shares an
empty volume with the Kaniko container, and extracted in the empty volume (only for local context).

### Limitations

* You cannot specify args for the Kaniko executor

#### Windows
* The docker-credential-wincred helper is not supported by Kaniko
