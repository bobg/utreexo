# Utreexo

This is an implementation of Utreexo,
a “dynamic accumulator” for utxos.
The original design is due to Tadge Dryja.
See [his talk at the 2019 MIT Bitcoin Expo](https://youtu.be/edRun-6ubCc).

In a conventional cryptocurrency network like Bitcoin,
all nodes on the network are obligated to maintain replicas of the “utxo set”:
the ever-changing set of valid, unspent coins and their denominations.
This allows each node independently to validate an attempt to spend a coin:
the node simply checks whether the coin’s unique ID exists in its copy of the utxo set.

But this means individual users can impose arbitrarily high costs on the entire network,
simply by creating large numbers of low-denomination coins that everyone must track.

The utreexo design places this tracking burden onto individual coin owners,
instead of onto everyone else ­
while still permitting everyone else to check the validity of spends.

In a blockchain based on the utreexo design,
nodes do not store all utxos.
Instead, they store the root hashes of *perfectly full Merkle binary trees* of utxos.

In a Merkle binary tree,
each node is either a leaf or has a hash computed from its children.
The root of such a tree thus _commits_ to the set of values it contains.
The contents of the tree can’t be changed without altering the root hash.

A perfectly full Merkle binary tree has 2^N leaves.
Any number of utxos can be represented with at most one 2^N tree for different values of N.
For instance, if there are 18 total utxos,
this can be represented with a 16-item perfectly full tree and a 2-item perfectly full tree.

Merkle trees permit the construction of _proofs of inclusion_ for the values represented in the tree.
The proof consists of the set of neighbors needed,
when climbing from a leaf toward the root,
to recreate the hashes along the way.
The proof must also indicate whether each neighbor is a lefthand or a righthand neighbor,
for proper hashing.
Checking such a proof amounts to calculating the chain of hashes and verifying that it ends up the same as the hash at the root of the tree.
As a simple example,
consider the Merkle tree with two leaves,
A and B,
and root hash R
(made from combining then hashing A and B).
The proof that A is in the tree is simply <B,right>.
This is enough information for a verifier to compute R and affirm it matches the root of the tree.
Similarly,
the proof that B is in the tree is <A,left>.

In a utreexo blockchain network,
transactions do not merely specify the unique ID of the utxo they wish to spend:
they must also specify its _proof_.
This allows nodes in the network both to check the validity of the utxo and to make the necessary changes to the Merkle trees involved.
Those changes involve removing the spent utxos and reordering things to maintain the invariant of zero-or-one perfectly full Merkle binary trees of each size.
At first glance this might seem impossible,
since the whole point of this design is that nodes no longer have to keep the full trees around,
only their root hashes.
As it happens,
all the necessary information for reconstructing affected parts of the trees is contained in the proofs that transactions supply when spending utxos.

The owners of the utxos do not merely take on the burden of storing their proofs:
they must also keep them up to date as the blockchain’s set of perfectly full Merkle binary trees undergoes continuous transformation.
Each update of the “utreexo” must therefore produce information that can be used to update those proofs,
and an owner must ensure all such updates are applied to a utxo’s proof before trying to spend it.
(However, as a service, a node may keep some number of those update records on hand,
allowing it to update any stale proofs it receives before trying to apply them to the current utreexo.)

See [Algorithm.md](Algorithm.md) for the workings of this implementation.
