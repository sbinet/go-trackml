# go-trackml

[![GoDoc](https://godoc.org/github.com/sbinet/go-trackml?status.svg)](https://godoc.org/github.com/sbinet/go-trackml)


`trackml` is a Go package to simplify working with the [High Energy Physics Tracking Machine Learning challenge][kaggle_trackml].

For more informations about the minute details of what `go-trackml` tries to do, please have a look at the Python version:

- https://github.com/LAL/trackml-library

`trackml` is a Go reimplementation of the above Python library.

## Installation

```sh
$> go get github.com/sbinet/go-trackml
```

## Documentation

Served by [GoDoc](https://godoc.org/github.com/sbinet/go-trackml).

## Example

```sh
$> go get github.com/sbinet/go-trackml/cmd/trkml-hough

$> trkml-hough -h
trkml-hough uses a Hough transform to make predictions.

Usage:

  $> trkml-hough [OPTIONS] <path-to-dataset> <evtid-prefix> [<path-to-test-dataset]

Examples:

  $> trkml-hough ./example_standard/dataset event000000200
  $> trkml-hough -npcus=+1 ./example_standard/dataset event000000200
  $> trkml-hough -npcus=-1 ./example_standard/dataset event000000200
  $> trkml-hough -npcus=-1 ./train_sample.zip event000001000

Options:

  -ncpus int
    	number of goroutines to use for the prediction (default 1)
  -prof-cpu
    	enable CPU profiling
  -prof-mem
    	enable MEM profiling
  -submit
     	create a submission file

$> ll example_standard/dataset/
total 56M
-rw-r--r-- 1 binet binet  13M Apr 25 18:36 event000000200-cells.csv
-rw-r--r-- 1 binet binet 4.2M Apr 25 18:36 event000000200-hits.csv
-rw-r--r-- 1 binet binet 915K Apr 25 18:36 event000000200-particles.csv
-rw-r--r-- 1 binet binet 9.5M Apr 25 18:36 event000000200-truth.csv
-rw-r--r-- 1 binet binet  14M Apr 25 18:36 event000000201-cells.csv
-rw-r--r-- 1 binet binet 4.5M Apr 25 18:36 event000000201-hits.csv
-rw-r--r-- 1 binet binet 967K Apr 25 18:36 event000000201-particles.csv
-rw-r--r-- 1 binet binet  10M Apr 25 18:36 event000000201-truth.csv

$> time trkml-hough ./example_standard/dataset event000000200
trkml-hough: loading [event000000200 from ./example_standard/dataset]...
trkml-hough: loading [event000000200 from ./example_standard/dataset]... [done]
trkml-hough: score for event 200: 0.1316012364071201
trkml-hough: loading the whole dataset "./example_standard/dataset"...
trkml-hough: score for event 200: 0.1316012364071201
trkml-hough: score for event 201: 0.1332602513710427
trkml-hough: loading the whole dataset "./example_standard/dataset"... [done]
trkml-hough: mean score: 0.13243074388908138

real  1m21.033s
user  1m22.541s
sys   0m0.569s
```

Compare to the Python version:

```sh
$> time python trkml-hough.py ./example_standard/dataset event000000200
   hit_id          x         y       z  volume_id  layer_id  module_id
0       1 -62.663200  -3.05090 -1502.5          7         2          1
1       2 -66.124702  -1.36730 -1502.5          7         2          1
2       3 -63.697701   1.73267 -1502.5          7         2          1
3       4 -82.501801 -14.09150 -1502.5          7         2          1
4       5 -74.343399   0.84469 -1502.5          7         2          1
Your score:  0.13153644878592863
Score for event 200: 0.132
Score for event 201: 0.133
Mean score: 0.132

real  7m12.351s
user  7m10.400s
sys   0m0.828s
```

### Going parallel

[Go](https://golang.org) has a few built-in facilities to apply concurrent programming.
The simple `trkml-hough` command leverages them:

```sh
$> time trkml-hough -ncpus=-1 ./example_standard/dataset event000000200
trkml-hough: loading [event000000200 from ./example_standard/dataset]...
trkml-hough: loading [event000000200 from ./example_standard/dataset]... [done]
trkml-hough: score for event 200: 0.13160123640712013
trkml-hough: loading the whole dataset "./example_standard/dataset"...
trkml-hough: score for event 200: 0.1316012364071201
trkml-hough: score for event 201: 0.13326025137104267
trkml-hough: loading the whole dataset "./example_standard/dataset"... [done]
trkml-hough: mean score: 0.13243074388908138

real  0m30.081s
user  1m26.658s
sys   0m0.741s
```

[cern]: https://home.cern
[lhc]: https://home.cern/topics/large-hadron-collider
[kaggle_trackml]: https://www.kaggle.com/c/trackml-particle-identification
