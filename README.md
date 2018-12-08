# kbuild

Build container images inside a Kubernetes Cluster with your local build context.

See [Kaniko](https://github.com/GoogleContainerTools/kaniko)

### Requirements
1. `~/.docker/config.json` exists and authenticated to a registry
2. [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl-binary-using-curl) is installed and properly configured 
3. a Kubernetes Cluster
4. a Container Registry

### Installation

For Linux or Mac use

```bash
curl -fsSL https://raw.githubusercontent.com/cedrickring/kbuild/scripts/get | bash
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

### How does kbuild work?

In order to use the local context, the context needs to be tar-ed, copied to an Init Container, which shares an
empty volume with the Kaniko container, and extracted in the empty volume. 

### Limitations

* Build args are not supported
* You cannot specify args for the Kaniko executor

#### Windows
* When running on windows, your `%TEMP%` directory must be in `C:` (or your default drive).
* The docker-credential-wincred helper is not supported by Kaniko