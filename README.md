Kagami is a general image board archiver. "Kagami" is Japanese for
"Mirror" (according to [Google Translate](https://translate.google.com/#en/ja/Mirror)), and that's exactly what
Kagami does - it mirrors existing Imageboards. It is published into
the public domain (see [LICENSE](./LICENSE)).

# Why use Kagami

- It has the potential to support a multitude of different board and
  engines
- It uses a lot of concurrency, which can speed things up a lot
  (mirroring a sample board needed 4 minutes and 23 seconds with
  minimal concurrency, ie. 4 threads, but only 22 seconds with the
  default settings)
- Instead of running an internal server, Kagami generates static
  output (except for searching)
- All metadata is saved into a single SQLite database
- Compiling generates a single binary, and hence does not require
  special runtimes or interpreters to be installed.

# Install 

After having made sure that go is installed, run:

```
$ git clone https://github.com/xatasan/kagami
$ go build -v
```

**Note:** the generated `kagami` binary shouldn't be installed
globally, because it need the `.tmpl` (go templates) files to generate
the mirror correctly.

# Run

Generally, executing Kagami takes the form of running:

```
$ ./kagami [host] [board]
```

while `[board]` is the hostname of the site, which must not be the
full host name (eg. instead of `8ch.net` one can also use `8ch` or
`8chan` as a synonym), while `[board]` doesn't need any slashes around
it.

As of now, the following boards/hosts are supported:

| `[host]` | type   | encodes |
|----------|--------|---------|
| 8ch      | board  | 8ch.net |
| 8chan    | board  | 8ch.net |

If nothing matches, it defaults to the vichan engine.

Kagami can be run once to mirror all the threads on the specified or
it can be set up to update itself periodically (for example with
`cron(8)`). 

## Important flags

- `-o`: Sets the directory to which the generated files should be
  written relative to. If not specified, it defaults to `./out/`. Your
  HTTP server should only have to host this directory.
- `-db`: Specifies path to database, which is used for catalog
  generation and saving metadata. Defaults to `./kagami.db`.
- `-d` and `-v`: enable debugging and verbose output
  respectively. Unless either is used, Kagami [generates no
  output](http://www.linfo.org/rule_of_silence.html).

Other flags, like `-D`, `-F` and `-W` regulate the number of threads
started for downloading threads, downloading files and writing threads
respectively.

# To do

- Improve catalog generator
- Add new Engines
- Make flags work for 8chan ([in
  progress](https://github.com/OpenIB/OpenIB/issues/192))
- Speed it up
- More flags
- Better defaults (ie. less need for flags)

# Changelog

## 0.0.0 (newest)

- It works
- It's hacky
- It is in need of rework

