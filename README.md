[![License](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go](https://github.com/sekiguchi-nagisa/guniset/actions/workflows/go.yml/badge.svg)](https://github.com/sekiguchi-nagisa/guniset/actions/workflows/go.yml)

# guniset

Unicode set extraction tool written in Go. It is heavily inspired by
[depp/uniset](https://github.com/depp/uniset)

## Usage

```sh
mkdir unicode_data   # unicode data directory
guniset download ./unicode_data  # download unicode data
GUNISET_DIR=./unicode_data guniset generate <set operation>
```

The ``GUNISET_DIR`` environmental variable must indicate a directory
having the following data

* ``DerivedGeneralCategory.txt``
* ``EastAsianWidth.txt``
* ``PropertyValueAliases.txt``
* ``Scripts.txt``
* ``ScriptExtensions.txt``
* ``PropList.txt``
* ``DerivedCoreProperties.txt``
* ``emoji-data.txt``
* ``DerivedBinaryProperties.txt``
* ``DerivedNormalizationProps.txt``
* ``GraphemeBreakProperty.txt``
* ``WordBreakProperty.txt``
* ``SentenceBreakProperty.txt``
* ``CaseFolding.txt``

## Set Operation

### Operators

* ``+``: union
* ``-``: difference
* ``*``: intersection
* ``!``: complement
* ``( )``: grouping
* ``@fold( )``: simple case folding

### Primitives

* ``cat:Cn,Me``: Unicode General Category set
* ``eaw:F,W``: East Asian Width set
* ``sc:Common``: Script set
* ``scx:Grek``: Script Extension set
* ``prop:White_Space``: Unicode property defined in ``PropList.txt``
* ``dcp:Grapheme_Base``: Unicode property defined in ``DerivedCoreProperties.txt``
* ``emoji:Emoji_Presentation``: Unicode property defined in ``emoji-data.txt``
* ``dbp:Bidi_Mirrored``: Unicode property defined in ``DerivedBinaryProperties.txt``
* ``dnp:FC_NFKC``: Unicode property defined in ``DerivedNormalizationProps.txt``
* ``gbp:Prepend``: Unicode property defined in ``GraphemeBreakProperty.txt``
* ``wbp:Extend``: Unicode property defined in ``WordBreakProperty.txt``
* ``sbp:Format``: Unicode property defined in ``SentenceBreakProperty.txt``
* ``U+1234``, ``0..1FFF``: Unicode code point

### Grammar

```
Expression 
    : UnionOrDiffEpxression

UnionOrDiffExpression 
    : IntersectionExpression ( ( '+' | '-' ) IntersectionExpression )*

IntersectionExpression
    : ComplementExpression ( '*' ComplementExpression )*

ComplementExpression
    : PrimaryExpression
    | '!' ComplementExpression
    | '@' 'fold' '(' Expression ')'

PrimaryExpression
    : ('cat' | 'gc') ':' CateList 
    | ('eaw' | 'ea') ':' EawList
    | 'sc' ':' PropList
    | 'scx' ':' PropList           # for script extensions
    | 'prop' ':' PropList          # for unicode properties
    | 'dcp' ':' PropList           # for derived core properties
    | 'emoji' ':' PropList         # for emoji
    | 'dbp' ':' PropList           # for derived binary properties
    | 'dnp' ':' PropList           # for derived normalization properties
    | 'gbp' ':' PropList           # for grapheme break properties
    | 'wbp' ':' PropList           # for word break properties
    | 'sbp' ':' PropList           # for sentence break properties
    | CodePoint '..' CodePoint
    | CodePoint
    | '(' Epxression ')'

CateList
    : Cate
    | Cate ',' CateList

Cate
    : 'Lu' | 'Ll' | 'Lt' | 'Lm' | 'Lo' | 'LC' | 'L'
    | 'Mn' | 'Mc' | 'Me' | 'M'
    | 'Nd' | 'Nl' | 'No' | 'N'
    | 'Pc' | 'Pd' | 'Ps' | 'Pe' | 'Pi' | 'Pf' | 'Po' | 'P'
    | 'Sm' | 'Sc' | 'Sk' | 'So' | 'S'
    | 'Zs' | 'Zl' | 'Zp' | 'Z'
    | 'Cc' | 'Cf' | 'Cs' | 'Co' | 'Cn' | 'C'
    | [a-zA-Z][a-zA-Z0-9_]+  # <other general category values and aliases>

EawList
    : Eaw
    | Eaw ',' EawList

Eaw
    : 'F' | 'W' | 'A' | 'Na' | 'N' | 'H'
    | [a-zA-Z][a-zA-Z0-9_]+  # <other east asian width aliases>

PropList
    : Prop
    | Prop ',' PropList

Prop
    : [a-zA-Z][a-zA-Z0-9_]+  # <other property names>

CodePoint
    : 'U+' [0-9a-fA-F]+
    | [0-9] [0-9a-fA-F]*
```
