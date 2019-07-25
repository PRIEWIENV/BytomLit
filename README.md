# Path Hashed Time Locked Contract (PHTLC)

PHTLC can be used in payment channel networks against wormhole attack.

## Introduction

### Hashed Time Locked Contract (HTLC)

Assume a collision-resistant hash function H and the condition R is chosen uniformly.
`HTLC (Alice, Bob, y, x, t)` is defined as: 
- If Bob produces the condition R such that H(R)=y before t days, Alice pays Bob x bitcoins.
- If t days elapse, Alice gets back x bitcoins.

### Payment Channel Networks (PCNs)

![](images/pcn.jpg)

### Wormhole Attack in PCNs

The wormhole attack in PCNs is first analyzed formally by [MMSK18](https://eprint.iacr.org/2018/472.pdf), in which the authors using a novel cryptographic primitive -- anonymous multi-hop locks (AMHLs) to address this problem.

![](images/wormhole_attack.png)

## PHTLC Scheme

Assume a collision-resistant hash function H and the condition R is chosen uniformly.
`PHTLC (Alice, Bob, y, x, t, m, U)` is defined as:

- If Bob produces the condition R such that H(R) = y and also provides aggregated signatures on message m signed by all the users in U before t days, Alice pays Bob x bitcoins.
- If t days elapse, Alice gets back x bitcoins.
