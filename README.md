[![License](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go](https://github.com/sekiguchi-nagisa/guniset/actions/workflows/go.yml/badge.svg)](https://github.com/sekiguchi-nagisa/guniset/actions/workflows/go.yml)

# guniset

Unicode set extraction tool written in Go. It is heavily inspired by 
[depp/uniset](https://github.com/depp/uniset)

## Usage
```sh
GUNISET_DIR=./unicode_data guniset <set operation>
```

The ``GUNISET_DIR`` environmental variable must indicate a directory 
having ``DerivedGeneralCategory.txt`` and ``EastAsianWidth.txt``

## Set Operation
### Operators
* ``+``: union
* ``-``: difference
* ``!``: complement
* ``( )``: grouping

### Primitives
* ``cat:Cn,Me``: Unicode General Category set
* ``eaw:F,W``: East Asian Width set
* ``U+1234``, ``0..1FFF``: Unicode code point

### Grammar
```
Expression 
    : UnionOrDiffEpxression

UnionOrDiffExpression 
    : ComplementExpression
    | UnionOrDiffExpression '+' ComplementExpression
    | UnionOrDiffExpression '-' ComplementExpression

ComplementExpression
    : PrimaryExpression
    | '!' ComplementExpression

PrimaryExpression
    : 'cat' ':' CateList 
    | 'eaw' ':' EawList
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

EawList
    : Eaw
    | Eaw ',' EawList

Eaw
    : 'F' | 'W' | 'A' | 'Na' | 'N' | 'H'
 
CodePoint
    : 'U+' [0-9a-fA-F]+
    | [0-9] [0-9a-fA-F]*
```
