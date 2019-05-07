# Utreexo algorithms and data structures

This document is an abstract specification for Utreexo algorithms and data structures.

**[NOTE: this is a draft.]**

## Terminology

An _n-Full Merkle Tree_ is one with 2^n leaves for some integer n>=0.

An _n-Full Merkle Root_ is the root hash of an n-Full Merkle Tree.

A _hasher_ is a function mapping an ordered pair of hashes to a new hash.

## Data structures

### Utreexo

A _Utreexo_ is a pair `<roots, hasher>`,
where `roots` is a _list of lists_ of Merkle Roots and `hasher` is a hasher.
More specifically, `roots[n]` is a list of n-Full Merkle Roots.
Except during updates of the Utreexo, each `roots[n]` has length zero or one.

Note: As described in this document, the leaves of a Utreexo are hashes.
These must be suitable as input to a Utreexo’s hasher but are otherwise unspecified
(and are not strictly required to _be_ hashes).

### Update record

An _update record_ is a set of `<hash, proofstep>` pairs,
where `proofstep` is a Merkle Proof Step.

### Merkle Proof Step

A _Merkle Proof Step_ is a pair `<hash, side>`,
where `hash` is a hash and `side` is one of `left` or `right`.

### Merkle Proof

A _Merkle Proof_ is an ordered sequence of Merkle Proof Steps plus `leaf`, a leaf value.

## Algorithms

### Verify proof

Inputs:
- `u`, a Utreexo,
- `proof`, a [Merkle Proof](#merkle-proof).

Outputs:
- a boolean: whether `proof` shows that `proof.leaf` is in an n-Full Merkle Tree for which `u` contains a root.

Procedure:
1. Let `n` be the number of steps in `proof`.
2. If `u.roots[n]` has length zero, return false.
3. Let `expected` be `u.roots[n][0]`.
4. Let `h` be `proof.leaf`.
5. For each step `s` in `proof`:
    1. If `s.side` is `left`, set `h´` to `u.hasher(s.hash, h)`.
    2. Otherwise, `s.side` is `right`. Set `h´` to `u.hasher(h, s.hash)`.
    3. Set `h` to `h´`.
6. Return `h == u.roots[n][0]`.

### Add and delete leaves

Inputs:
- `u`, a Utreexo,
- `deletions`, a list of Merkle proofs for leaves to be deleted,
- `insertions`, a list of leaf values to insert.

Outputs:
- an update record compactly describing the changes to the Utreexo.

Notes:
- Implementations should return with `u` unchanged in case of failure.
- Some proofs in `deletions` may be only partially verified against `u` by this procedure.
  Implementations may prefer to perform complete [verification](#verify-proof)
  before permitting deletion.

Procedure:
1. Let `r` be a new, empty [update record](#update-record).
2. For each proof `p` in `deletions`:
    1. Let `height` be 0.
    2. Let `n` be the number of steps in `p`.
    3. Let `h` be `p.leaf`.
    4. Repeat the following:
        1. If `h` is in `u.roots[height]`, remove it
           (preserving the order of remaining elements in `u.roots[height]`)
           and exit the loop.
        2. If `height >= n` fail (`p` fails verification).
        3. Let `s` be `p.steps[height]`.
        4. Append `s.hash` to `u.roots[height]`.
        5. If `s.side` is `left`, set `hash´` to `u.hasher(s.hash, h)`.
        6. Otherwise, `s.side` is `right`. Set `hash´` to `u.hasher(h, s.hash)`.
        7. Set `hash` to `hash´`.
        8. Set `height` to `height+1`.
3. For each hash `h` in `insertions`:
    1. Append `h` to `u.roots[0]`.
4. _(Coalescing.)_ For each index `index` in `[0, u.roots.length)`:
    1. While `u.roots[index].length > 1`:
        1. Let `a` and `b` be the last two elements of `u.roots[index]`.
        2. Decrease `u.roots[index].length` by 2.
        3. Let `h` be `u.hasher(a, b)`.
        4. Append `h` to `u.roots[index+1]`.
        5. Let `sl` be a new [Merkle Proof Step](#merkle-proof-step) with `hash = a` and `side = left`.
        6. Let `sr` be a new Merkle Proof Step with `hash = b` and `side = right`.
        7. Add `<left, sr>` to r.
        8. Add `<right, sl>` to r.
5. Return `r`.

### Construct proof

After a new leaf is added to a Utreexo,
it is possible to construct a Merkle Proof for that value
using the Utreexo,
the update record that resulted from the insertion,
and the leaf value itself.

Input:
- `u`, a Utreexo,
- `r`, an [update record](#update-record),
- `leaf`, a leaf value.

Output:
- a [Merkle Proof](#merkle-proof).

Procedure:

1. Let `p` be a new `Merkle Proof` with an empty list of `steps` and `p.leaf` set to `leaf`.
2. Let `h` be `leaf`.
2. Repeat the following:
    1. If `r.nodes` does not contain the pair `<h, s>` for some step `s`, return `p`.
    2. Otherwise, append `s` to `p.steps`.
    3. If `s.side` is `left`, set `h´` to `u.hasher(s.hash, h)`.
    4. Otherwise, `s.side` is `right`. Set `h´` to `u.hasher(h, s.hash)`.
    5. Set `h` to `h´`.

### Update proof

After each update of a Utreexo (the deletion or addition leaves),
existing Merkle Proofs must be updated to remain valid.

Input:
- `u`, a Utreexo,
- `r`, an [update record](#update-record),
- `p`, a [Merkle Proof](#merkle-proof).

Procedure:

1. Let `h` be `p.leaf`.
2. Let `i` be 0.
3. Repeat the following:
    1. If `i` > `p.steps.length`, fail (invalid proof).
    2. If `u.roots[i]` is non-empty and `u.roots[i][0]` is `h`,
       then set `p.steps.length` to `i` and return.
    3. If `r` contains the pair `<h, s>` for some step `s`, then:
        1. Set `p.steps.length` to `i`.
        2. Append `s` to `p.steps`.
    4. Otherwise (if `r` does not contain the pair `<h, s>`):
        1. If `i == p.steps.length`, fail (invalid proof).
        2. Otherwise, let `s` be `p.steps[i]`.
    5. If `s.side` is `left`, set `h´` to `u.hasher(s.hash, h)`.
    6. Otherwise, `s.side` is `right`. Set `h´` to `u.hasher(h, s.hash)`.
    7. Set `h` to `h´`.
