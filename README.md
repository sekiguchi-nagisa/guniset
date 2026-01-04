[![License](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go](https://github.com/sekiguchi-nagisa/guniset/actions/workflows/go.yml/badge.svg)](https://github.com/sekiguchi-nagisa/guniset/actions/workflows/go.yml)

# guniset

Unicode set extraction tool written in Go. It is heavily inspired by
[depp/uniset](https://github.com/depp/uniset)

## Usage

```sh
GUNISET_DIR=./unicode_data guniset generate <set operation>
```

The ``GUNISET_DIR`` environmental variable must indicate a directory
having the following data

* ``DerivedGeneralCategory.txt``
* ``EastAsianWidth.txt``
* ``PropertyValueAliases.txt``
* ``Scripts.txt``
* ``ScriptExtensions.txt``

## Set Operation

### Operators

* ``+``: union
* ``-``: difference
* ``*``: intersection
* ``!``: complement
* ``( )``: grouping

### Primitives

* ``cat:Cn,Me``: Unicode General Category set
* ``eaw:F,W``: East Asian Width set
* ``sc:Common``: Script set
* ``scx:Grek``: Script Extension set
* ``U+1234``, ``0..1FFF``: Unicode code point

### Grammar

```
Expression 
    : UnionOrDiffEpxression

UnionOrDiffExpression 
    : IntersectionExpression
    | UnionOrDiffExpression '+' IntersectionExpression
    | UnionOrDiffExpression '-' IntersectionExpression

IntersectionExpression
    : ComplementExpression
    | IntersectionExpression '*' ComplementExpression

ComplementExpression
    : PrimaryExpression
    | '!' ComplementExpression

PrimaryExpression
    : ('cat' | 'gc') ':' CateList 
    | ('eaw' | 'ea') ':' EawList
    | 'sc' ':' ScriptList
    | 'scx' ':' ScriptList        # for script extensions
    | CodePoint '..' CodePoint
    | CodePoint
    | '(' Epxression ')'

CateList
    : Cate
    | Cate ',' CateList

Cate
    : 'Lu' | 'Ll' | 'Lt' | 'Lm' | 'Lo'
    | 'Mn' | 'Mc' | 'Me'
    | 'Nd' | 'Nl' | 'No'
    | 'Pc' | 'Pd' | 'Ps' | 'Pe' | 'Pi' | 'Pf' | 'Po'
    | 'Sm' | 'Sc' | 'Sk' | 'So' | 'Zs' | 'Zl' | 'Zp'
    | 'Cc' | 'Cf' | 'Cs' | 'Co' | 'Cn'
    | <other general category values and aliases>

EawList
    : Eaw
    | Eaw ',' EawList

Eaw
    : 'F' | 'W' | 'A' | 'Na' | 'N' | 'H'
    | <other east asian width aliases>

ScriptList
    : Sctipt
    | Script ',' ScriptList

Script
    : <script values and aliases>

CodePoint
    : 'U+' [0-9a-fA-F]+
    | [0-9] [0-9a-fA-F]*
```
