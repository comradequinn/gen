# Installing Gen

To install `gen`, download the appropriate tarball for your `os` from the [releases](https://github.com/comradequinn/gen/releases/) page. Extract the binary and place it somewhere accessible to your `$PATH` variable.

Optionally, you can use the below script to do that for you

```bash
export VERSION="v1.3.9"; export OS="linux-amd64"; wget "https://github.com/comradequinn/gen/releases/download/${VERSION}/gen-${VERSION}-${OS}.tar.gz" && tar -xf "gen-${VERSION}-${OS}.tar.gz" && rm -f "gen-${VERSION}-${OS}.tar.gz" && chmod +x gen && sudo mv gen /usr/local/bin/
```

## Authentication

You can configure `gen` to access the `Gemini API` either via the publicly available `Generative Language API` endpoints (as used by `Google AI Studio`) or via a `Vertex AI` endpoint managed within a `Google Cloud Platform (GCP)` project.

### Generative Language API (Google AI Studio)

To use `gen` via the `Generative Language API`, set and export your `Gemini API Key` as the conventional environment variable for that value: `GEMINI_API_KEY`.

If you do not already have a `Gemini API Key`, they are available free from [Google AI Studio](https://aistudio.google.com), [here](https://aistudio.google.com/apikey).

For convenience, you may wish to add the envar definition to your `~/.bashrc` file. An example of doing this is shown below.

```bash
# file: ~/.bashrc

export GEMINI_API_KEY="myPriVatEApI_keY_1234567890"
```

Remember that you will need to open a new terminal or `source` the `~/.bashrc` file for the above to take effect.

Once this is done, `gen` will default to using the `Generative Language API` and your `GEMINI_API_KEY` unless you explicitly specify `Vertex AI (Google Cloud Platform)` credentials to use instead; in which case they will take precedence.

#### Enabling Sudo

To enable `gen` to be used with `sudo`, the `Gemini API Key` must be passed explicitly, rather than inferred from the `GEMINI_API_KEY` envar. To do this directly, run `gen` with the `--access-token` argument (or `-a`) and pass the value of the envar as shown below.

```bash
gen -a "$GEMINI_API_KEY" -x "create the directory /etc/temp-gen"
# >> executing... [mkdir /etc/temp-gen]
# >> Error: mkdir: cannot create directory ‘/etc/temp-gen’: Permission denied
```

As shown, the above fails with a permission error. However, prefixing it with `sudo` in the usual manner will allow `gen's` privileges to be escalated (_as long as the `GEMINI_API_KEY` is passed explicitly_). This is shown below.

```bash
sudo gen -a "$GEMINI_API_KEY" -x "create the directory /etc/temp-gen"
# >> [sudo] password for user: ********
# >> executing... [mkdir /etc/temp-gen]
# >> OK
```

To enable `sudo` support by default, as with any command executed via a `posix` compliant shell, `aliases` can be used to assign defaults arguments. An example is shown below that configures `gen` to run in the above manner, by default.

```bash
# file: ~/.bashrc

export GEMINI_API_KEY="myPriVatEApI_keY_1234567890"

alias sudo='sudo ' # ensure sudo is aliased to enable subsequent alias expansion
alias gen="gen --access-token \"$GEMINI_API_KEY\""
```

> Note that the alias is not configured to embed `sudo` directly, only to allow `sudo` to be effective when specified. It is not advisable to run `gen` in `sudo` at all times, for the usual reasons of security and system protection.

With the above alias configured, `gen` can be be run with `sudo` without specifying the `access-token`. As shown below.

```bash
sudo gen -c -x "remove the directory"
# >> [sudo] password for user: ********
# >> executing... [rmdir /etc/temp-gen]
# >> OK
```

### Vertex AI (Google Cloud Platform)

To use `gen` with a `Vertex AI` `Gemini API` endpoint, firstly configure `ADC (application default credentials)` on your workstation, if you have not already done so, by running the below.

```bash
gcloud auth application-default login --disable-quota-project
```

You can then render `access tokens` using `gcloud auth application-default print-access-token`. These can be passed to `gen` using an `--access-token` (or `-a`) argument. A `GCP Project` and a `GCS Bucket` must also be specified, using `--gcp-project` (or `-p`) and `--gcs-bucket` (or `-b`), respectively. An example is shown below.

```bash
gen --access-token "$(gcloud auth application-default print-access-token)" --gcp-project "my-project" --gcs-bucket "my-bucket" "what is the weather like in London tomorrow?"
```

When you specify `Vertex AI` credentials, they take precedence over any `GEMINI_API_KEY` you may have set to authenticate with the `Generative Language API`.

### Configuring Defaults

As with any command executed via a `posix` compliant shell, `aliases` can be used to assign defaults arguments. An example is shown below that configures a default `access-token` and `gcp-project`.

```bash
# file: ~/.bashrc

alias gen='gen --access-token "$(gcloud auth application-default print-access-token)" --gcp-project "my-project" --gcs-bucket "my-bucket"'
```

Users of the shell can then simply run `gen` directly and implicitly use those `gcp credentials`. As shown below.

```bash
gen "what is the weather like in London tomorrow?"
```

## Removal

To remove `gen`, delete the binary from `/usr/bin` (or the location it was originally installed to). You may also wish to delete its application directory. This stores user preferences and session history and is located at `~/.gen`.