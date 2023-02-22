# CSVQL 

<a href="https://github.com/adrianolaselva/csvql"><img align="right" src="./docs/img/logo.png" alt="csvql" title="csvql" width="160px"/></a>


![github actions](https://github.com/adrianolaselva/csvql/actions/workflows/build.yml/badge.svg)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/adrianolaselva/csvql/badges/quality-score.png?b=main)](https://scrutinizer-ci.com/g/adrianolaselva/csvql/?branch=main)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/adrianolaselva/csvql/badges/quality-score.png?b=master)](https://scrutinizer-ci.com/g/adrianolaselva/csvql/?branch=master)
[![GoDoc](https://godoc.org/github.com/adrianolaselva/csvql?status.svg)](https://pkg.go.dev/github.com/adrianolaselva/csvql)
![GitHub issues](https://img.shields.io/github/issues/adrianolaselva/csvql)
![license](http://img.shields.io/badge/license-Apache%20v2-blue.svg)

CLI tool developed in GO to facilitate the handling of CSV files, making it possible to import large files locally
and manipulate them through sqlite-based SQL statements.

This tool's main objective is to provide a way to manipulate large csv files locally, facilitating analyzes that 
require the use of tools such as excel.

## Features

**Current:**

- Import `.csv` file for manipulation.
- Using sqlite-based SQL statements.

**future features:**

- Export `queryes` in `.csv`, `.json`, `.jsonl` or `sqlite3`.

## Installation

Run the command below to download and install the latest version of the tool.

```sh
curl -s "https://raw.githubusercontent.com/adrianolaselva/csvql/main/bin/install" | bash
```
> Install tool from the `latest` version.

**Note: To install from a specific version, just pass the release number in the url**

```sh
curl -s "https://raw.githubusercontent.com/adrianolaselva/csvql/v1.0.0/bin/install" | bash
```
> Install tool from the `v1.0.0` version.

**Note: Soon you can also choose to download the binary install and use it**

## Uninstallation

```sh
curl -s "https://raw.githubusercontent.com/adrianolaselva/csvql/main/bin/install" | bash
```
> Uninstall tool.

## Usage

Once installed, just run the command below passing a CSV file as a parameter through the `-f` flag and the delimiter 
used through the `-d` flag.

```sh
csvql run -f test.csv -d ";"
```
> Example initializing a file named `test.csv` using `;` as delimiter.

**Example:**

Below is an example of how the tool works, importing a csv file delimited by the character `;`.

```shell
csvql> select origin_id, description, metric_value, metric_date from rows limit 10;
origin_id    description                    metric_value   metric_date  
1007549851   Amazon Sales Revenue           0,35           01/02/2023   
1007549852   Bahia Sales Revenue            0,21           01/02/2023   
1007683973   Ceará Sales Revenue            0,65           01/02/2023   
1007710146   Espírito Santo Sales Revenue   0,58           01/02/2023   
1007772105   Goiás Sales Revenue            0,06           01/02/2023   
1007778716   Maranhão Sales Revenue         0,65           01/02/2023   
1007780734   Mato Grosso Sales Revenue      0,23           01/02/2023   
1007789224   São Paulo Sales Revenue        0,48           01/02/2023   
1007975972   Tocantins Sales Revenue        3,01           01/02/2023   
1008060883   Rio de Janeiro Sales Revenue   0,39           01/02/2023
```
> Example of SQL execution after loading `.csv` file.

## References

- [sqlite database](https://www.tutorialspoint.com/sqlite/index.htm)

## License

table is released under the MIT License (Expat). See the [full license](https://github.com/adrianolaselva/table/blob/main/license).