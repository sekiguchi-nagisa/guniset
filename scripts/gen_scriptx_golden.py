#!/usr/bin/env python3
"""Generate golden files for Script_Extensions (scx) property.

Unlike the other generators (which list the raw property values), this script
computes the FULL value of the `scx:X` set operation as implemented by
`LoadScriptXMap` (op/load.go), independently from the guniset binary:

  raw = union of ranges whose ScriptExtensions.txt value list contains X
  if X is Common or Inherited:
      scx(X) = Script(X) - (union of all listed ranges in ScriptExtensions.txt)
  elif X is Unknown (Zzzz):
      # no Scripts.txt entry for unassigned code points: the fallback in
      # FillScriptUnknown (op/load.go) yields the complement of all scripts
      scx(X) = complement of all code points listed in Scripts.txt
  else:
      scx(X) = Script(X) | raw

Usage:
  python3 scripts/gen_scriptx_golden.py <output_dir>

The script downloads Scripts.txt, ScriptExtensions.txt and PropertyValueAliases.txt
of the revision specified by UNICODE_VERSION directly from unicode.org (using only
the standard library), and writes `<output_dir>/scx_<abbr>.test` / `.golden` pairs.
"""

import os
import re
import sys
import urllib.request

# Unicode version of the UCD data to download (keep in sync with
# scripts/get_unicode_data.sh and the test/generate/unicode16* directories).
UNICODE_VERSION = "16.0.0"
UCD_BASE_URL = f"https://www.unicode.org/Public/{UNICODE_VERSION}/ucd"


def fetch_ucd_file(filename):
    """Download a UCD data file of UNICODE_VERSION and return its text content."""
    url = f"{UCD_BASE_URL}/{filename}"
    req = urllib.request.Request(url, headers={"User-Agent": "gen_scriptx_golden"})
    with urllib.request.urlopen(req) as resp:
        return resp.read().decode("utf-8")

ENTRY_RE = re.compile(
    r"^(?P<first>[0-9A-F]+)(?:\.\.(?P<last>[0-9A-F]+))?\s*;\s*(?P<props>.+?)(?:\s*#.*)?$"
)


def parse_entries(text, filename):
    """Parse the content of a UCD range file into (first, last, value) tuples."""
    entries = []
    for line in text.splitlines():
        line = line.split("#", 1)[0].rstrip()
        if not line:
            continue
        m = ENTRY_RE.match(line)
        if not m:
            raise RuntimeError(f"failed to parse {filename}: {line}")
        first = int(m.group("first"), 16)
        last = int(m.group("last"), 16) if m.group("last") else first
        entries.append((first, last, m.group("props")))
    return entries


def minimize(ranges):
    """Sort and merge adjacent/overlapping ranges."""
    result = []
    for first, last in sorted(ranges):
        if result and first <= result[-1][1] + 1:
            result[-1] = (result[-1][0], max(result[-1][1], last))
        else:
            result.append((first, last))
    return result


def subtract(ranges, cut):
    """Subtract the set of ranges `cut` from `ranges`."""
    result = list(ranges)
    for c_first, c_last in cut:
        nxt = []
        for first, last in result:
            if c_last < first or last < c_first:  # no overlap
                nxt.append((first, last))
                continue
            if first < c_first:
                nxt.append((first, c_first - 1))
            if c_last < last:
                nxt.append((c_last + 1, last))
        result = nxt
    return minimize(result)


def union(*range_lists):
    result = []
    for ranges in range_lists:
        result.extend(ranges)
    return minimize(result)


def main():
    if len(sys.argv) != 2:
        print(f"usage: {sys.argv[0]} <output_dir>", file=sys.stderr)
        return 1
    output_dir = os.path.realpath(sys.argv[1])
    os.makedirs(output_dir, exist_ok=True)

    # abbr -> long script name (e.g. Latn -> Latin)
    # PropertyValueAliases.txt is a plain `value; alias; ...` file, not a range file
    abbr_to_long = {}
    for line in fetch_ucd_file("PropertyValueAliases.txt").splitlines():
        line = line.split("#", 1)[0].strip()
        if not line:
            continue
        ss = [s.strip() for s in line.split(";")]
        if ss[0] != "sc" or len(ss) < 3:
            continue
        abbr_to_long[ss[1]] = ss[2]

    # Script property: script long name -> ranges
    sc_ranges = {}
    for first, last, props in parse_entries(fetch_ucd_file("Scripts.txt"), "Scripts.txt"):
        script = props.strip()
        sc_ranges.setdefault(script, []).append((first, last))

    # Script_Extensions property
    # scx_raw: script abbr -> ranges listed for the script
    # found: all ranges listed in the file (for Common/Inherited subtraction)
    scx_raw = {}
    found = []
    for first, last, props in parse_entries(
        fetch_ucd_file("ScriptExtensions.txt"), "ScriptExtensions.txt"
    ):
        found.append((first, last))
        for abbr in props.split():
            scx_raw.setdefault(abbr, []).append((first, last))

    # every script appearing in ScriptExtensions.txt must be known
    for abbr in sorted(scx_raw):
        if abbr not in abbr_to_long:
            raise RuntimeError(f"unknown script in ScriptExtensions.txt: {abbr}")

    # output cases: scripts listed in ScriptExtensions.txt + special cases
    targets = sorted(scx_raw, key=lambda a: a.lower())
    specials = ["Common", "Inherited", "Unknown"]

    for target in targets + specials:
        if target in abbr_to_long:
            long = abbr_to_long[target]
        elif target in sc_ranges:
            long = target
        elif target == "Unknown":
            long = "Unknown"  # Zzzz has no single-code-point entry in Scripts.txt
        else:
            raise RuntimeError(f"script not found in Scripts.txt: {target}")
        sc = minimize(sc_ranges.get(long, []))
        if target == "Unknown":
            # no ScriptExtensions.txt entry and no Scripts.txt entry: the set is
            # the complement of all scripts listed in Scripts.txt (see FillScriptUnknown)
            known = union(*[minimize(v) for v in sc_ranges.values()])
            result = subtract([(0x0000, 0x10FFFF)], known)
        elif target in ("Common", "Inherited"):
            result = subtract(sc, minimize(found))
        else:
            result = union(sc, minimize(scx_raw[target]))
        # file names use the abbr, op uses the long name (e.g. scx_latn.test = "scx:Latin")
        base = "scx_" + target.lower()
        with open(os.path.join(output_dir, base + ".test"), "w") as f:
            f.write(f"scx:{long}\n")
        with open(os.path.join(output_dir, base + ".golden"), "w") as f:
            for first, last in result:
                f.write(f"{{ 0x{first:04X}, 0x{last:04X} }},\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())