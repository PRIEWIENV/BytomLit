# BytomLit Specifications

The specifications are currently a work-in-progress.

## Introduction

This is version 0.

Lightning is a protocol for making fast payments with Bitcoin using a
network of channels.

### Channels

Lightning works by establishing *channels*: two participants create a
Lightning payment channel that contains some amount of bitcoin (e.g.,
0.1 bitcoin) that they've locked up on the Bitcoin network. It is
spendable only with both their signatures.

Initially they each hold a bitcoin transaction that sends all the
bitcoin (e.g. 0.1 bitcoin) back to one party.  They can later sign a new bitcoin
transaction that splits these funds differently, e.g. 0.09 bitcoin to one
party, 0.01 bitcoin to the other, and invalidate the previous bitcoin
transaction so it won't be spent.

### Conditional Payments

A Lightning channel only allows payment between two participants, but channels can be connected together to form a network that allows payments between all members of the network. This requires the technology of a conditional payment, which can be added to a channel,
e.g. "you get 0.01 bitcoin if you reveal the secret within 6 hours".
Once the recipient presents the secret, that bitcoin transaction is
replaced with one lacking the conditional payment and adding the funds
to that recipient's output.

See [BOLT #2: Adding an HTLC](02-peer-protocol.md#adding-an-htlc-update_add_htlc) for the commands a participant uses to add a conditional payment, and [BOLT #3: Commitment Transaction](03-transactions.md#commitment-transaction) for the
complete format of the bitcoin transaction.

### Forwarding

Such a conditional payment can be safely forwarded to another
participant with a lower time limit, e.g. "you get 0.01 bitcoin if you reveal the secret
within 5 hours".  This allows channels to be chained into a network
without trusting the intermediaries.

See [BOLT #2: Forwarding HTLCs](02-peer-protocol.md#forwarding-htlcs) for details on forwarding payments, [BOLT #4: Packet Structure](04-onion-routing.md#packet-structure) for how payment instructions are transported.

### Network Topology

To make a payment, a participant needs to know what channels it can
send through.  Participants tell each other about channel and node
creation, and updates.

See [BOLT #7: P2P Node and Channel Discovery](07-routing-gossip.md)
for details on the communication protocol, and [BOLT #10: DNS
Bootstrap and Assisted Node Location](10-dns-bootstrap.md) for initial
network bootstrap.

## Glossary and Terminology Guide

* #### *Announcement*:
   * A gossip message sent between *[peers](#peers)* intended to aid the discovery of a *[channel](#channel)* or a *[node](#node)*.

* #### `chain_hash`:
   * The uniquely identifying hash of the target blockchain (usually the genesis hash).
     This allows *[nodes](#node)* to create and reference *channels* on
     several blockchains. Nodes are to ignore any messages that reference a
     `chain_hash` that are unknown to them. Unlike `bitcoin-cli`, the hash is
     not reversed but is used directly.

     For the main chain Bitcoin blockchain, the `chain_hash` value MUST be
     (encoded in hex):
     `6fe28c0ab6f1b372c1a6a246ae63f74f931e8365e15a089c68d6190000000000`.

* #### *Channel*:
   * A fast, off-chain method of mutual exchange between two *[peers](#peers)*.
   To transact funds, peers exchange signatures to create an updated *[commitment transaction](#commitment-transaction)*.
   * _See closure methods: [mutual close](#mutual-close), [revoked transaction close](#revoked-transaction-close), [unilateral close](#unilateral-close)_
   * _See related: [route](#route)_

* #### *Closing transaction*:
   * A transaction generated as part of a *[mutual close](#mutual-close)*. A closing transaction is similar to a _commitment transaction_, but with no pending payments.
   * _See related: [commitment transaction](#commitment-transaction), [funding transaction](#funding-transaction), [penalty transaction](#penalty-transaction)_

* #### *Commitment number*:
   * A 48-bit incrementing counter for each *[commitment transaction](#commitment-transaction)*; counters
    are independent for each *peer* in the *channel* and start at 0.
   * _See container: [commitment transaction](#commitment-transaction)_
   * _See related: [closing transaction](#closing-transaction), [funding transaction](#funding-transaction), [penalty transaction](#penalty-transaction)_

* #### *Commitment revocation private key*:
   * Every *[commitment transaction](#commitment-transaction)* has a unique commitment revocation private-key
    value that allows the other *peer* to spend all outputs
    immediately: revealing this key is how old commitment
    transactions are revoked. To support revocation, each output of the
    commitment transaction refers to the commitment revocation public key.
   * _See container: [commitment transaction](#commitment-transaction)_
   * _See originator: [per-commitment secret](#per-commitment-secret)_

* #### *Commitment transaction*:
   * A transaction that spends the *[funding transaction](#funding-transaction)*.
   Each *peer* holds the other peer's signature for this transaction, so that each
   always has a commitment transaction that it can spend. After a new
   commitment transaction is negotiated, the old one is *revoked*.
   * _See parts: [commitment number](#commitment-number), [commitment revocation private key](#commitment-revocation-private-key), [HTLC](#HTLC-Hashed-Time-Locked-Contract), [per-commitment secret](#per-commitment-secret), [outpoint](#outpoint)_
   * _See related: [closing transaction](#closing-transaction), [funding transaction](#funding-transaction), [penalty transaction](#penalty-transaction)_
   * _See types: [revoked commitment transaction](#revoked-commitment-transaction)_

* #### *Final node*:
   * The final recipient of a packet that is routing a payment from an *[origin node](#origin-node)* through some number of *[hops](#hop)*. It is also the final *[receiving peer](#receiving-peer)* in a chain.
   * _See category: [node](#node)_
   * _See related: [origin node](#origin-node), [processing node](#processing-node)_

* #### *Funding transaction*:
   * An irreversible on-chain transaction that pays to both *[peers](#peers)* on a *[channel](#channel)*.
   It can only be spent by mutual consent.
   * _See related: [closing transaction](#closing-transaction), [commitment transaction](#commitment-transaction), [penalty transaction](#penalty-transaction)_

* #### *Hop*:
   * A *[node](#node)*. Generally, an intermediate node lying between an *[origin node](#origin-node)* and a *[final node](#final-node)*.
   * _See category: [node](#node)_

* #### *HTLC*: Hashed Time Locked Contract.
   * A conditional payment between two *[peers](#peers)*: the recipient can spend
    the payment by presenting its signature and a *payment preimage*,
    otherwise the payer can cancel the contract by spending it after
    a given time. These are implemented as outputs from the
    *[commitment transaction](#commitment-transaction)*.
   * _See container: [commitment transaction](#commitment-transaction)_
   * _See parts: [Payment hash](#Payment-hash), [Payment preimage](#Payment-preimage)_

* #### *Invoice*: A request for funds on the Lightning Network, possibly
    including payment type, payment amount, expiry, and other
    information. This is how payments are made on the Lightning
    Network, rather than using Bitcoin-style addresses.

* #### *It's ok to be odd*:
   * A rule applied to some numeric fields that indicates either optional or
     compulsory support for features. Even numbers indicate that both endpoints
     MUST support the feature in question, while odd numbers indicate
     that the feature MAY be disregarded by the other endpoint.

* #### *MSAT*:
   * A millisatoshi, often used as a field name.

* #### *Mutual close*:
   * A cooperative close of a *[channel](#channel)*, accomplished by broadcasting an unconditional
    spend of the *[funding transaction](#funding-transaction)* with an output to each *peer*
    (unless one output is too small, and thus is not included).
   * _See related: [revoked transaction close](#revoked-transaction-close), [unilateral close](#unilateral-close)_

* #### *Node*:
   * A computer or other device that is part of the Lightning network.
   * _See related: [peers](#peers)_
   * _See types: [final node](#final-node), [hop](#hop), [origin node](#origin-node), [processing node](#processing-node), [receiving node](#receiving-node), [sending node](#sending-node)_

* #### *Origin node*:
   * The *[node](#node)* that originates a packet that will route a payment through some number of [hops](#hop) to a *[final node](#final-node)*. It is also the first [sending peer](#sending-peer) in a chain.
   * _See category: [node](#node)_
   * _See related: [final node](#final-node), [processing node](#processing-node)_

* #### *Outpoint*:
  * A transaction hash and output index that uniquely identify an unspent transaction output. Needed to compose a new transaction, as an input.
  * _See related: [funding transaction](#funding-transaction), [commitment transaction](#commitment-transaction)_

* #### *Payment hash*:
   * The *[HTLC](#HTLC-Hashed-Time-Locked-Contract)* contains the payment hash, which is the hash of the
    *[payment preimage](#Payment-preimage)*.
   * _See container: [HTLC](#HTLC-Hashed-Time-Locked-Contract)_
   * _See originator: [Payment preimage](#Payment-preimage)_

* #### *Payment preimage*:
   * Proof that payment has been received, held by
    the final recipient, who is the only person who knows this
    secret. The final recipient releases the preimage in order to
    release funds. The payment preimage is hashed as the *[payment hash](#Payment-hash)*
    in the *[HTLC](#HTLC-Hashed-Time-Locked-Contract)*.
   * _See container: [HTLC](#HTLC-Hashed-Time-Locked-Contract)_
   * _See derivation: [payment hash](#Payment-hash)_

* #### *Peers*:
   * Two *[nodes](#node)* that are in communication with each other.
      * Two peers may gossip with each other prior to setting up a channel.
      * Two peers may establish a *[channel](#channel)* through which they transact.
   * _See related: [node](#node)_

* #### *Penalty transaction*:
   * A transaction that spends all outputs of a *[revoked commitment
    transaction](#revoked-commitment-transaction)*, using the *commitment revocation private key*. A *[peer](#peers)* uses this
    if the other peer tries to "cheat" by broadcasting a *[revoked commitment
    transaction](#revoked-commitment-transaction)*.
   * _See related: [closing transaction](#closing-transaction), [commitment transaction](#commitment-transaction), [funding transaction](#funding-transaction)_

* #### *Per-commitment secret*:
   * Every *[commitment transaction](#commitment-transaction)* derives its keys from a per-commitment secret,
     which is generated such that the series of per-commitment secrets
     for all previous commitments can be stored compactly.
   * _See container: [commitment transaction](#commitment-transaction)_
   * _See derivation: [commitment revocation private key](#commitment-revocation-private-key)_

* #### *Processing node*:
   * A *[node](#node)* that is processing a packet that originated with an *[origin node](#origin-node)* and that is being sent toward a *[final node](#final-node)* in order to route a payment. It acts as a *[receiving peer](#receiving-peer)* to receive the message, then a [sending peer](#sending-peer) to send on the packet.
   * _See category: [node](#node)_
   * _See related: [final node](#final-node), [origin node](#origin-node)_

* #### *Receiving node*:
   * A *[node](#node)* that is receiving a message.
   * _See category: [node](#node)_
   * _See related: [sending node](#sending-node)_

* #### *Receiving peer*:
   * A *[node](#node)* that is receiving a message from a directly connected *peer*.
   * _See category: [peer](#Peers)_
   * _See related: [sending peer](#sending-peer)_

* #### *Revoked commitment transaction*:
   * An old *[commitment transaction](#commitment-transaction)* that has been revoked because a new commitment transaction has been negotiated.
   * _See category: [commitment transaction](#commitment-transaction)_

* #### *Revoked transaction close*:
   * An invalid close of a *[channel](#channel)*, accomplished by broadcasting a *revoked
    commitment transaction*. Since the other *peer* knows the
    *commitment revocation secret key*, it can create a *[penalty transaction](#penalty-transaction)*.
   * _See related: [mutual close](#mutual-close), [unilateral close](#unilateral-close)_

* #### *Route*: 
  * A path across the Lightning Network that enables a payment
    from an *origin node* to a *[final node](#final-node)* across one or more
    *[hops](#hop)*.
  * _See related: [channel](#channel)_

* #### *Sending node*:
   * A *[node](#node)* that is sending a message.
   * _See category: [node](#node)_
   * _See related: [receiving node](#receiving-node)_

* #### *Sending peer*:
   * A *[node](#node)* that is sending a message to a directly connected *peer*.
   * _See category: [peer](#Peers)_
   * _See related: [receiving peer](#receiving-peer)_.

* #### *Unilateral close*:
   * An uncooperative close of a *[channel](#channel)*, accomplished by broadcasting a
    *[commitment transaction](#commitment-transaction)*. This transaction is larger (i.e. less
    efficient) than a *[closing transaction](#closing-transaction)*, and the *[peer](#peers)* whose
    commitment is broadcast cannot access its own outputs for some
    previously-negotiated duration.
   * _See related: [mutual close](#mutual-close), [revoked transaction close](#revoked-transaction-close)_

## Base Protocol

This protocol assumes an underlying authenticated and ordered transport mechanism that takes care of framing individual messages.
[BOLT #8](08-transport.md) specifies the canonical transport layer used in Lightning, though it can be replaced by any transport that fulfills the above guarantees.

The default TCP port is 9735. This corresponds to hexadecimal `0x2607`: the Unicode code point for LIGHTNING.

All data fields are unsigned big-endian unless otherwise specified.

### Connection Handling and Multiplexing

Implementations MUST use a single connection per peer; channel messages (which include a channel ID) are multiplexed over this single connection.

### Message Format

All messages are of the form:

1. `type`: a 2-byte big-endian field indicating the type of message
2. `payload`: a variable-length payload that comprises the remainder of
   the message and that conforms to a format matching the `type`

The `type` field indicates how to interpret the `payload` field.
The format for each individual type is defined by a specification in this repository.
The type follows the _it's ok to be odd_ rule, so nodes MAY send _odd_-numbered types without ascertaining that the recipient understands it.

A sending node:
  - MUST NOT send an evenly-typed message not listed here without prior negotiation.

A receiving node:
  - upon receiving a message of _odd_, unknown type:
    - MUST ignore the received message.
  - upon receiving a message of _even_, unknown type:
    - MUST fail the channels.

The messages are grouped logically into four groups, ordered by the most significant bit that is set:

  - Setup & Control (types `0`-`31`): messages related to connection setup, control, supported features, and error reporting (described below)
  - Channel (types `32`-`127`): messages used to setup and tear down micropayment channels (described in [BOLT #2](02-peer-protocol.md))
  - Commitment (types `128`-`255`): messages related to updating the current commitment transaction, which includes adding, revoking, and settling HTLCs as well as updating fees and exchanging signatures (described in [BOLT #2](02-peer-protocol.md))
  - Routing (types `256`-`511`): messages containing node and channel announcements, as well as any active route exploration (described in [BOLT #7](07-routing-gossip.md))

The size of the message is required by the transport layer to fit into a 2-byte unsigned int; therefore, the maximum possible size is 65535 bytes.

A node:
  - MUST ignore any additional data within a message beyond the length that it expects for that type.
  - upon receiving a known message with insufficient length for the contents:
    - MUST fail the channels.
  - that negotiates an option in this specification:
    - MUST include all the fields annotated with that option.

### Rationale

By default `SHA2` and Bitcoin public keys are both encoded as
big endian, thus it would be unusual to use a different endian for
other fields.

Length is limited to 65535 bytes by the cryptographic wrapping, and
messages in the protocol are never more than that length anyway.

The _it's ok to be odd_ rule allows for future optional extensions
without negotiation or special coding in clients. The "ignore
additional data" rule similarly allows for future expansion.

Implementations may prefer to have message data aligned on an 8-byte
boundary (the largest natural alignment requirement of any type here);
however, adding a 6-byte padding after the type field was considered
wasteful: alignment may be achieved by decrypting the message into
a buffer with 6-bytes of pre-padding.

## Type-Length-Value Format

Throughout the protocol, a TLV (Type-Length-Value) format is used to allow for
the backwards-compatible addition of new fields to existing message types.

A `tlv_record` represents a single field, encoded in the form:

* [`varint`: `type`]
* [`varint`: `length`]
* [`length`: `value`]

A `varint` is a variable-length, unsigned integer encoding using the
[BigSize](#appendix-a-bigsize-test-vectors) format, which resembles the bitcoin
CompactSize encoding but uses big-endian for multi-byte values rather than
little-endian. 

A `tlv_stream` is a series of (possibly zero) `tlv_record`s, represented as the
concatenation of the encoded `tlv_record`s. When used to extend existing
messages, a `tlv_stream` is typically placed after all currently defined fields.

The `type` is a varint encoded using the BigSize format. It functions as a
message-specific, 64-bit identifier for the `tlv_record` determining how the
contents of `value` should be decoded.

The `length` is a varint encoded using the BigSize format signaling the size of
`value` in bytes.

The `value` depends entirely on the `type`, and should be encoded or decoded
according to the message-specific format determined by `type`.

### Requirements

The sending node:
 - MUST order `tlv_record`s in a `tlv_stream` by monotonically-increasing `type`.
 - MUST minimally encode `type` and `length`.
 - SHOULD NOT use redundant, variable-length encodings in a `tlv_record`.

The receiving node:
 - if zero bytes remain before parsing a `type`:
   - MUST stop parsing the `tlv_stream`.
 - if a `type` or `length` is not minimally encoded:
   - MUST fail to parse the `tlv_stream`.
 - if decoded `type`s are not monotonically-increasing:
   - MUST fail to parse the `tlv_stream`.
 - if `length` exceeds the number of bytes remaining in the message:
   - MUST fail to parse the `tlv_stream`.
 - if `type` is known:
   - MUST decode the next `length` bytes using the known encoding for `type`.
   - if `length` is not exactly equal to that required for the known encoding for `type`:
     - MUST fail to parse the `tlv_stream`.
   - if variable-length fields within the known encoding for `type` are not minimal:
     - MUST fail to parse the `tlv_stream`.
 - otherwise, if `type` is unknown:
   - if `type` is even:
     - MUST fail to parse the `tlv_stream`.
   - otherwise, if `type` is odd:
     - MUST discard the next `length` bytes.

### Rationale

The primary advantage in using TLV is that a reader is able to ignore new fields
that it does not understand, since each field carries the exact size of the
encoded element. Without TLV, even if a node does not wish to use a particular
field, the node is forced to add parsing logic for that field in order to
determine the offset of any fields that follow.

The monotonicity constraint ensures that all `type`s are unique and can appear
at most once. Fields that map to complex objects, e.g. vectors, maps, or
structs, should do so by defining the encoding such that the object is
serialized within a single `tlv_record`. The uniqueness constraint, among other
things, enables the following optimizations:
 - canonical ordering is defined independent of the encoded `value`s.
 - canonical ordering can be known at compile-time, rather that being determined
   dynamically at the time of encoding.
 - verifying canonical ordering requires less state and is less-expensive.
 - variable-size fields can reserve their expected size up front, rather than
   appending elements sequentially and incurring double-and-copy overhead.

The use of a varint for `type` and `length` permits a space savings for small
`type`s or short `value`s. This potentially leaves more space for application
data over the wire or in an onion payload.

All `type`s must appear in increasing order to create a canonical encoding of
the underlying `tlv_record`s. This is crucial when computing signatures over a
`tlv_stream`, as it ensures verifiers will be able to recompute the same message
digest as the signer. Note that the canonical ordering over the set of fields
can be enforced even if the verifier does not understand what the fields
contain.

Writers should avoid using redundant, variable-length encodings in a
`tlv_record` since this results in encoding the length twice and complicates
computing the outer length. As an example, when writing a variable length byte
array, the `value` should contain only the raw bytes and forgo an additional
internal length since the `tlv_record` already carries the number of bytes that
follow. On the other hand, if a `tlv_record` contains multiple, variable-length
elements then this would not be considered redundant, and is needed to allow the
receiver to parse individual elements from `value`.

## Fundamental Types

Various fundamental types are referred to in the message specifications:

* `byte`: an 8-bit byte
* `u16`: a 2 byte unsigned integer
* `u32`: a 4 byte unsigned integer
* `u64`: an 8 byte unsigned integer

Inside TLV records which contain a single value, leading zeros in
integers can be omitted:

* `tu16`: a 0 to 2 byte unsigned integer
* `tu32`: a 0 to 4 byte unsigned integer
* `tu64`: a 0 to 8 byte unsigned integer

The following convenience types are also defined:

* `chain_hash`: a 32-byte chain identifier (see [BOLT #0](00-introduction.md#glossary-and-terminology-guide))
* `channel_id`: a 32-byte channel_id (see [BOLT #2](02-peer-protocol.md#definition-of-channel-id)
* `sha256`: a 32-byte SHA2-256 hash
* `signature`: a 64-byte bitcoin Elliptic Curve signature
* `point`: a 33-byte Elliptic Curve point (compressed encoding as per [SEC 1 standard](http://www.secg.org/sec1-v2.pdf#subsubsection.2.3.3))
* `short_channel_id`: an 8 byte value identifying a channel (see [BOLT #7](07-routing-gossip.md#definition-of-short-channel-id))

## Setup Messages

### The `init` Message

Once authentication is complete, the first message reveals the features supported or required by this node, even if this is a reconnection.

[BOLT #9](09-features.md) specifies lists of global and local features. Each feature is generally represented in `globalfeatures` or `localfeatures` by 2 bits. The least-significant bit is numbered 0, which is _even_, and the next most significant bit is numbered 1, which is _odd_.

Both fields `globalfeatures` and `localfeatures` MUST be padded to bytes with 0s.

1. type: 16 (`init`)
2. data:
   * [`u16`:`gflen`]
   * [`gflen*byte`:`globalfeatures`]
   * [`u16`:`lflen`]
   * [`lflen*byte`:`localfeatures`]

The 2-byte `gflen` and `lflen` fields indicate the number of bytes in the immediately following field.

#### Requirements

The sending node:
  - MUST send `init` as the first Lightning message for any connection.
  - MUST set feature bits as defined in [BOLT #9](09-features.md).
  - MUST set any undefined feature bits to 0.
  - SHOULD use the minimum lengths required to represent the feature fields.

The receiving node:
  - MUST wait to receive `init` before sending any other messages.
  - MUST respond to known feature bits as specified in [BOLT #9](09-features.md).
  - upon receiving unknown _odd_ feature bits that are non-zero:
    - MUST ignore the bit.
  - upon receiving unknown _even_ feature bits that are non-zero:
    - MUST fail the connection.

#### Rationale

This semantic allows both future incompatible changes and future backward compatible changes. Bits should generally be assigned in pairs, in order that optional features may later become compulsory.

Nodes wait for receipt of the other's features to simplify error
diagnosis when features are incompatible.

The feature masks are split into local features (which only affect the
protocol between these two nodes) and global features (which can affect
HTLCs and are thus also advertised to other nodes).

### The `error` Message

For simplicity of diagnosis, it's often useful to tell a peer that something is incorrect.

1. type: 17 (`error`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`u16`:`len`]
   * [`len*byte`:`data`]

The 2-byte `len` field indicates the number of bytes in the immediately following field.

#### Requirements

The channel is referred to by `channel_id`, unless `channel_id` is 0 (i.e. all bytes are 0), in which case it refers to all channels.

The funding node:
  - for all error messages sent before (and including) the `funding_created` message:
    - MUST use `temporary_channel_id` in lieu of `channel_id`.

The fundee node:
  - for all error messages sent before (and not including) the `funding_signed` message:
    - MUST use `temporary_channel_id` in lieu of `channel_id`.

A sending node:
  - when sending `error`:
    - MUST fail the channel referred to by the error message.
  - SHOULD send `error` for protocol violations or internal errors that make channels unusable or that make further communication unusable.
  - SHOULD send `error` with the unknown `channel_id` in reply to messages of type `32`-`255` related to unknown channels.
  - MAY send an empty `data` field.
  - when failure was caused by an invalid signature check:
    - SHOULD include the raw, hex-encoded transaction in reply to a `funding_created`, `funding_signed`, `closing_signed`, or `commitment_signed` message.
  - when `channel_id` is 0:
    - MUST fail all channels with the receiving node.
    - MUST close the connection.
  - MUST set `len` equal to the length of `data`.

The receiving node:
  - upon receiving `error`:
    - MUST fail the channel referred to by the error message, if that channel is with the sending node.
  - if no existing channel is referred to by the message:
    - MUST ignore the message.
  - MUST truncate `len` to the remainder of the packet (if it's larger).
  - if `data` is not composed solely of printable ASCII characters (For reference: the printable character set includes byte values 32 through 126, inclusive):
    - SHOULD NOT print out `data` verbatim.

#### Rationale

There are unrecoverable errors that require an abort of conversations;
if the connection is simply dropped, then the peer may retry the
connection. It's also useful to describe protocol violations for
diagnosis, as this indicates that one peer has a bug.

It may be wise not to distinguish errors in production settings, lest
it leak information — hence, the optional `data` field.

## Control Messages

### The `ping` and `pong` Messages

In order to allow for the existence of long-lived TCP connections, at
times it may be required that both ends keep alive the TCP connection at the
application level. Such messages also allow obfuscation of traffic patterns.

1. type: 18 (`ping`)
2. data:
    * [`u16`:`num_pong_bytes`]
    * [`u16`:`byteslen`]
    * [`byteslen*byte`:`ignored`]

The `pong` message is to be sent whenever a `ping` message is received. It
serves as a reply and also serves to keep the connection alive, while
explicitly notifying the other end that the receiver is still active. Within
the received `ping` message, the sender will specify the number of bytes to be
included within the data payload of the `pong` message.

1. type: 19 (`pong`)
2. data:
    * [`u16`:`byteslen`]
    * [`byteslen*byte`:`ignored`]

#### Requirements

A node sending a `ping` message:
  - SHOULD set `ignored` to 0s.
  - MUST NOT set `ignored` to sensitive data such as secrets or portions of initialized
memory.
  - if it doesn't receive a corresponding `pong`:
    - MAY terminate the network connection,
      - and MUST NOT fail the channels in this case.
  - SHOULD NOT send `ping` messages more often than once every 30 seconds.

A node sending a `pong` message:
  - SHOULD set `ignored` to 0s.
  - MUST NOT set `ignored` to sensitive data such as secrets or portions of initialized
 memory.

A node receiving a `ping` message:
  - SHOULD fail the channels if it has received significantly in excess of one `ping` per 30 seconds.
  - if `num_pong_bytes` is less than 65532:
    - MUST respond by sending a `pong` message, with `byteslen` equal to `num_pong_bytes`.
  - otherwise (`num_pong_bytes` is **not** less than 65532):
    - MUST ignore the `ping`.

A node receiving a `pong` message:
  - if `byteslen` does not correspond to any `ping`'s `num_pong_bytes` value it has sent:
    - MAY fail the channels.

### Rationale

The largest possible message is 65535 bytes; thus, the maximum sensible `byteslen`
is 65531 — in order to account for the type field (`pong`) and the `byteslen` itself. This allows
a convenient cutoff for `num_pong_bytes` to indicate that no reply should be sent.

Connections between nodes within the network may be long lived, as payment
channels have an indefinite lifetime. However, it's likely that
no new data will be
exchanged for a
significant portion of a connection's lifetime. Also, on several platforms it's possible that Lightning
clients will be put to sleep without prior warning. Hence, a
distinct `ping` message is used, in order to probe for the liveness of the connection on
the other side, as well as to keep the established connection active.

Additionally, the ability for a sender to request that the receiver send a
response with a particular number of bytes enables nodes on the network to
create _synthetic_ traffic. Such traffic can be used to partially defend
against packet and timing analysis — as nodes can fake the traffic patterns of
typical exchanges without applying any true updates to their respective
channels.

When combined with the onion routing protocol defined in
[BOLT #4](04-onion-routing.md),
careful statistically driven synthetic traffic can serve to further bolster the
privacy of participants within the network.

Limited precautions are recommended against `ping` flooding, however some
latitude is given because of network delays. Note that there are other methods
of incoming traffic flooding (e.g. sending _odd_ unknown message types, or padding
every message maximally).

Finally, the usage of periodic `ping` messages serves to promote frequent key
rotations as specified within [BOLT #8](08-transport.md).

## 2: Peer Protocol for Channel Management

The peer channel protocol has three phases: establishment, normal
operation, and closing.

### Channel

#### Definition of `channel_id`

Some messages use a `channel_id` to identify the channel. It's
derived from the funding transaction by combining the `funding_txid`
and the `funding_output_index`, using big-endian exclusive-OR
(i.e. `funding_output_index` alters the last 2 bytes).

Prior to channel establishment, a `temporary_channel_id` is used,
which is a random nonce.

Note that as duplicate `temporary_channel_id`s may exist from different
peers, APIs which reference channels by their channel id before the funding
transaction is created are inherently unsafe. The only protocol-provided
identifier for a channel before funding_created has been exchanged is the
(source_node_id, destination_node_id, temporary_channel_id) tuple. Note that
any such APIs which reference channels by their channel id before the funding
transaction is confirmed are also not persistent - until you know the script
pubkey corresponding to the funding output nothing prevents duplicative channel
ids.


#### Channel Establishment

After authenticating and initializing a connection ([BOLT #8](08-transport.md)
and [BOLT #1](01-messaging.md#the-init-message), respectively), channel establishment may begin.
This consists of the funding node (funder) sending an `open_channel` message,
followed by the responding node (fundee) sending `accept_channel`. With the
channel parameters locked in, the funder is able to create the funding
transaction and both versions of the commitment transaction, as described in
[BOLT #3](03-transactions.md#bolt-3-bitcoin-transaction-and-script-formats).
The funder then sends the outpoint of the funding output with the `funding_created`
message, along with the signature for the fundee's version of the commitment
transaction. Once the fundee learns the funding outpoint, it's able to
generate the signature for the funder's version of the commitment transaction and send it
over using the `funding_signed` message.

Once the channel funder receives the `funding_signed` message, it
must broadcast the funding transaction to the Bitcoin network. After
the `funding_signed` message is sent/received, both sides should wait
for the funding transaction to enter the blockchain and reach the
specified depth (number of confirmations). After both sides have sent
the `funding_locked` message, the channel is established and can begin
normal operation. The `funding_locked` message includes information
that will be used to construct channel authentication proofs.


        +-------+                              +-------+
        |       |--(1)---  open_channel  ----->|       |
        |       |<-(2)--  accept_channel  -----|       |
        |       |                              |       |
        |   A   |--(3)--  funding_created  --->|   B   |
        |       |<-(4)--  funding_signed  -----|       |
        |       |                              |       |
        |       |--(5)--- funding_locked  ---->|       |
        |       |<-(6)--- funding_locked  -----|       |
        +-------+                              +-------+

        - where node A is 'funder' and node B is 'fundee'

If this fails at any stage, or if one node decides the channel terms
offered by the other node are not suitable, the channel establishment
fails.

Note that multiple channels can operate in parallel, as all channel
messages are identified by either a `temporary_channel_id` (before the
funding transaction is created) or a `channel_id` (derived from the
funding transaction).

##### The `open_channel` Message

This message contains information about a node and indicates its
desire to set up a new channel. This is the first step toward creating
the funding transaction and both versions of the commitment transaction.

1. type: 32 (`open_channel`)
2. data:
   * [`chain_hash`:`chain_hash`]
   * [`32*byte`:`temporary_channel_id`]
   * [`u64`:`funding_satoshis`]
   * [`u64`:`push_msat`]
   * [`u64`:`dust_limit_satoshis`]
   * [`u64`:`max_htlc_value_in_flight_msat`]
   * [`u64`:`channel_reserve_satoshis`]
   * [`u64`:`htlc_minimum_msat`]
   * [`u32`:`feerate_per_kw`]
   * [`u16`:`to_self_delay`]
   * [`u16`:`max_accepted_htlcs`]
   * [`point`:`funding_pubkey`]
   * [`point`:`revocation_basepoint`]
   * [`point`:`payment_basepoint`]
   * [`point`:`delayed_payment_basepoint`]
   * [`point`:`htlc_basepoint`]
   * [`point`:`first_per_commitment_point`]
   * [`byte`:`channel_flags`]
   * [`u16`:`shutdown_len`] (`option_upfront_shutdown_script`)
   * [`shutdown_len*byte`:`shutdown_scriptpubkey`] (`option_upfront_shutdown_script`)

The `chain_hash` value denotes the exact blockchain that the opened channel will
reside within. This is usually the genesis hash of the respective blockchain.
The existence of the `chain_hash` allows nodes to open channels
across many distinct blockchains as well as have channels within multiple
blockchains opened to the same peer (if it supports the target chains).

The `temporary_channel_id` is used to identify this channel on a per-peer basis until the
funding transaction is established, at which point it is replaced
by the `channel_id`, which is derived from the funding transaction.

`funding_satoshis` is the amount the sender is putting into the
channel. `push_msat` is an amount of initial funds that the sender is
unconditionally giving to the receiver. `dust_limit_satoshis` is the
threshold below which outputs should not be generated for this node's
commitment or HTLC transactions (i.e. HTLCs below this amount plus
HTLC transaction fees are not enforceable on-chain). This reflects the
reality that tiny outputs are not considered standard transactions and
will not propagate through the Bitcoin network. `channel_reserve_satoshis`
is the minimum amount that the other node is to keep as a direct
payment. `htlc_minimum_msat` indicates the smallest value HTLC this
node will accept.

`max_htlc_value_in_flight_msat` is a cap on total value of outstanding
HTLCs, which allows a node to limit its exposure to HTLCs; similarly,
`max_accepted_htlcs` limits the number of outstanding HTLCs the other
node can offer.

`feerate_per_kw` indicates the initial fee rate in satoshi per 1000-weight
(i.e. 1/4 the more normally-used 'satoshi per 1000 vbytes') that this
side will pay for commitment and HTLC transactions, as described in
[BOLT #3](03-transactions.md#fee-calculation) (this can be adjusted
later with an `update_fee` message).

`to_self_delay` is the number of blocks that the other node's to-self
outputs must be delayed, using `OP_CHECKSEQUENCEVERIFY` delays; this
is how long it will have to wait in case of breakdown before redeeming
its own funds.

`funding_pubkey` is the public key in the 2-of-2 multisig script of
the funding transaction output.

The various `_basepoint` fields are used to derive unique
keys as described in [BOLT #3](03-transactions.md#key-derivation) for each commitment
transaction. Varying these keys ensures that the transaction ID of
each commitment transaction is unpredictable to an external observer,
even if one commitment transaction is seen; this property is very
useful for preserving privacy when outsourcing penalty transactions to
third parties.

`first_per_commitment_point` is the per-commitment point to be used
for the first commitment transaction,

Only the least-significant bit of `channel_flags` is currently
defined: `announce_channel`. This indicates whether the initiator of
the funding flow wishes to advertise this channel publicly to the
network, as detailed within [BOLT #7](07-routing-gossip.md#bolt-7-p2p-node-and-channel-discovery).

The `shutdown_scriptpubkey` allows the sending node to commit to where
funds will go on mutual close, which the remote node should enforce
even if a node is compromised later.

[ FIXME: Describe dangerous feature bit for larger channel amounts. ]

##### The `accept_channel` Message

This message contains information about a node and indicates its
acceptance of the new channel. This is the second step toward creating the
funding transaction and both versions of the commitment transaction.

1. type: 33 (`accept_channel`)
2. data:
   * [`32*byte`:`temporary_channel_id`]
   * [`u64`:`dust_limit_satoshis`]
   * [`u64`:`max_htlc_value_in_flight_msat`]
   * [`u64`:`channel_reserve_satoshis`]
   * [`u64`:`htlc_minimum_msat`]
   * [`u32`:`minimum_depth`]
   * [`u16`:`to_self_delay`]
   * [`u16`:`max_accepted_htlcs`]
   * [`point`:`funding_pubkey`]
   * [`point`:`revocation_basepoint`]
   * [`point`:`payment_basepoint`]
   * [`point`:`delayed_payment_basepoint`]
   * [`point`:`htlc_basepoint`]
   * [`point`:`first_per_commitment_point`]
   * [`u16`:`shutdown_len`] (`option_upfront_shutdown_script`)
   * [`shutdown_len*byte`:`shutdown_scriptpubkey`] (`option_upfront_shutdown_script`)

##### The `funding_created` Message

This message describes the outpoint which the funder has created for
the initial commitment transactions. After receiving the peer's
signature, via `funding_signed`, it will broadcast the funding transaction.

1. type: 34 (`funding_created`)
2. data:
    * [`32*byte`:`temporary_channel_id`]
    * [`sha256`:`funding_txid`]
    * [`u16`:`funding_output_index`]
    * [`signature`:`signature`]


##### The `funding_signed` Message

This message gives the funder the signature it needs for the first
commitment transaction, so it can broadcast the transaction knowing that funds
can be redeemed, if need be.

This message introduces the `channel_id` to identify the channel. It's derived from the funding transaction by combining the `funding_txid` and the `funding_output_index`, using big-endian exclusive-OR (i.e. `funding_output_index` alters the last 2 bytes).

1. type: 35 (`funding_signed`)
2. data:
    * [`channel_id`:`channel_id`]
    * [`signature`:`signature`]

##### The `funding_locked` Message

This message indicates that the funding transaction has reached the `minimum_depth` asked for in `accept_channel`. Once both nodes have sent this, the channel enters normal operating mode.

1. type: 36 (`funding_locked`)
2. data:
    * [`channel_id`:`channel_id`]
    * [`point`:`next_per_commitment_point`]

#### Channel Close

Nodes can negotiate a mutual close of the connection, which unlike a
unilateral close, allows them to access their funds immediately and
can be negotiated with lower fees.

Closing happens in two stages:
1. one side indicates it wants to clear the channel (and thus will accept no new HTLCs)
2. once all HTLCs are resolved, the final channel close negotiation begins.

        +-------+                              +-------+
        |       |--(1)-----  shutdown  ------->|       |
        |       |<-(2)-----  shutdown  --------|       |
        |       |                              |       |
        |       | <complete all pending HTLCs> |       |
        |   A   |                 ...          |   B   |
        |       |                              |       |
        |       |--(3)-- closing_signed  F1--->|       |
        |       |<-(4)-- closing_signed  F2----|       |
        |       |              ...             |       |
        |       |--(?)-- closing_signed  Fn--->|       |
        |       |<-(?)-- closing_signed  Fn----|       |
        +-------+                              +-------+

##### Closing Initiation: `shutdown`

Either node (or both) can send a `shutdown` message to initiate closing,
along with the `scriptpubkey` it wants to be paid to.

1. type: 38 (`shutdown`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`u16`:`len`]
   * [`len*byte`:`scriptpubkey`]


##### Closing Negotiation: `closing_signed`

Once shutdown is complete and the channel is empty of HTLCs, the final
current commitment transactions will have no HTLCs, and closing fee
negotiation begins.  The funder chooses a fee it thinks is fair, and
signs the closing transaction with the `scriptpubkey` fields from the
`shutdown` messages (along with its chosen fee) and sends the signature;
the other node then replies similarly, using a fee it thinks is fair.  This
exchange continues until both agree on the same fee or when one side fails
the channel.

1. type: 39 (`closing_signed`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`u64`:`fee_satoshis`]
   * [`signature`:`signature`]

#### Normal Operation

Once both nodes have exchanged `funding_locked` (and optionally [`announcement_signatures`](07-routing-gossip.md#the-announcement_signatures-message)), the channel can be used to make payments via Hashed Time Locked Contracts.

Changes are sent in batches: one or more `update_` messages are sent before a
`commitment_signed` message, as in the following diagram:

        +-------+                               +-------+
        |       |--(1)---- update_add_htlc ---->|       |
        |       |--(2)---- update_add_htlc ---->|       |
        |       |<-(3)---- update_add_htlc -----|       |
        |       |                               |       |
        |       |--(4)--- commitment_signed --->|       |
        |   A   |<-(5)---- revoke_and_ack ------|   B   |
        |       |                               |       |
        |       |<-(6)--- commitment_signed ----|       |
        |       |--(7)---- revoke_and_ack ----->|       |
        |       |                               |       |
        |       |--(8)--- commitment_signed --->|       |
        |       |<-(9)---- revoke_and_ack ------|       |
        +-------+                               +-------+

Counter-intuitively, these updates apply to the *other node's*
commitment transaction; the node only adds those updates to its own
commitment transaction when the remote node acknowledges it has
applied them via `revoke_and_ack`.

Thus each update traverses through the following states:

1. pending on the receiver
2. in the receiver's latest commitment transaction
3. ... and the receiver's previous commitment transaction has been revoked,
   and the update is pending on the sender
4. ... and in the sender's latest commitment transaction
5. ... and the sender's previous commitment transaction has been revoked


As the two nodes' updates are independent, the two commitment
transactions may be out of sync indefinitely. This is not concerning:
what matters is whether both sides have irrevocably committed to a
particular update or not (the final state, above).

##### Forwarding HTLCs

In general, a node offers HTLCs for two reasons: to initiate a payment of its own,
or to forward another node's payment. In the forwarding case, care must
be taken to ensure the *outgoing* HTLC cannot be redeemed unless the *incoming*
HTLC can be redeemed. The following requirements ensure this is always true.

The respective **addition/removal** of an HTLC is considered *irrevocably committed* when:

1. The commitment transaction **with/without** it is committed to by both nodes, and any
previous commitment transaction **without/with** it has been revoked, OR
2. The commitment transaction **with/without** it has been irreversibly committed to
the blockchain.

##### `cltv_expiry_delta` Selection

Once an HTLC has timed out, it can either be fulfilled or timed-out;
care must be taken around this transition, both for offered and received HTLCs.

Consider the following scenario, where A sends an HTLC to B, who
forwards to C, who delivers the goods as soon as the payment is
received.

1. C needs to be sure that the HTLC from B cannot time out, even if B becomes
   unresponsive; i.e. C can fulfill the incoming HTLC on-chain before B can
   time it out on-chain.

2. B needs to be sure that if C fulfills the HTLC from B, it can fulfill the
   incoming HTLC from A; i.e. B can get the preimage from C and fulfill the incoming
   HTLC on-chain before A can time it out on-chain.

The critical settings here are the `cltv_expiry_delta` in
[BOLT #7](07-routing-gossip.md#the-channel_update-message) and the
related `min_final_cltv_expiry` in [BOLT #11](11-payment-encoding.md#tagged-fields).
`cltv_expiry_delta` is the minimum difference in HTLC CLTV timeouts, in
the forwarding case (B). `min_final_cltv_expiry` is the minimum difference
between HTLC CLTV timeout and the current block height, for the
terminal case (C).

Note that a node is at risk if it accepts an HTLC in one channel and
offers an HTLC in another channel with too small of a difference between
the CLTV timeouts.  For this reason, the `cltv_expiry_delta` for the
*outgoing* channel is used as the delta across a node.

The worst-case number of blocks between outgoing and
incoming HTLC resolution can be derived, given a few assumptions:

* a worst-case reorganization depth `R` blocks
* a grace-period `G` blocks after HTLC timeout before giving up on
  an unresponsive peer and dropping to chain
* a number of blocks `S` between transaction broadcast and the
  transaction being included in a block

The worst case is for a forwarding node (B) that takes the longest
possible time to spot the outgoing HTLC fulfillment and also takes
the longest possible time to redeem it on-chain:

1. The B->C HTLC times out at block `N`, and B waits `G` blocks until
   it gives up waiting for C. B or C commits to the blockchain,
   and B spends HTLC, which takes `S` blocks to be included.
2. Bad case: C wins the race (just) and fulfills the HTLC, B only sees
   that transaction when it sees block `N+G+S+1`.
3. Worst case: There's reorganization `R` deep in which C wins and
   fulfills. B only sees transaction at `N+G+S+R`.
4. B now needs to fulfill the incoming A->B HTLC, but A is unresponsive: B waits `G` more
   blocks before giving up waiting for A. A or B commits to the blockchain.
5. Bad case: B sees A's commitment transaction in block `N+G+S+R+G+1` and has
   to spend the HTLC output, which takes `S` blocks to be mined.
6. Worst case: there's another reorganization `R` deep which A uses to
   spend the commitment transaction, so B sees A's commitment
   transaction in block `N+G+S+R+G+R` and has to spend the HTLC output, which
   takes `S` blocks to be mined.
7. B's HTLC spend needs to be at least `R` deep before it times out,
   otherwise another reorganization could allow A to timeout the
   transaction.

Thus, the worst case is `3R+2G+2S`, assuming `R` is at least 1. Note that the
chances of three reorganizations in which the other node wins all of them is
low for `R` of 2 or more. Since high fees are used (and HTLC spends can use
almost arbitrary fees), `S` should be small; although, given that block times are
irregular and empty blocks still occur, `S=2` should be considered a
minimum. Similarly, the grace period `G` can be low (1 or 2), as nodes are
required to timeout or fulfill as soon as possible; but if `G` is too low it increases the
risk of unnecessary channel closure due to networking delays.

There are four values that need be derived:

1. the `cltv_expiry_delta` for channels, `3R+2G+2S`: if in doubt, a
   `cltv_expiry_delta` of 12 is reasonable (R=2, G=1, S=2).

2. the deadline for offered HTLCs: the deadline after which the channel has to be failed
   and timed out on-chain. This is `G` blocks after the HTLC's
   `cltv_expiry`: 1 block is reasonable.

3. the deadline for received HTLCs this node has fulfilled: the deadline after which
the channel has to be failed and the HTLC fulfilled on-chain before its
   `cltv_expiry`. See steps 4-7 above, which imply a deadline of `2R+G+S`
   blocks before `cltv_expiry`: 7 blocks is reasonable.

4. the minimum `cltv_expiry` accepted for terminal payments: the
   worst case for the terminal node C is `2R+G+S` blocks (as, again, steps
   1-3 above don't apply). The default in
   [BOLT #11](11-payment-encoding.md) is 9, which is slightly more
   conservative than the 7 that this calculation suggests.

##### Adding an HTLC: `update_add_htlc`

Either node can send `update_add_htlc` to offer an HTLC to the other,
which is redeemable in return for a payment preimage. Amounts are in
millisatoshi, though on-chain enforcement is only possible for whole
satoshi amounts greater than the dust limit (in commitment transactions these are rounded down as
specified in [BOLT #3](03-transactions.md)).

The format of the `onion_routing_packet` portion, which indicates where the payment
is destined, is described in [BOLT #4](04-onion-routing.md).

1. type: 128 (`update_add_htlc`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`u64`:`id`]
   * [`u64`:`amount_msat`]
   * [`sha256`:`payment_hash`]
   * [`u32`:`cltv_expiry`]
   * [`1366*byte`:`onion_routing_packet`]

##### Removing an HTLC: `update_fulfill_htlc`, `update_fail_htlc`, and `update_fail_malformed_htlc`

For simplicity, a node can only remove HTLCs added by the other node.
There are four reasons for removing an HTLC: the payment preimage is supplied,
it has timed out, it has failed to route, or it is malformed.

To supply the preimage:

1. type: 130 (`update_fulfill_htlc`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`u64`:`id`]
   * [`32*byte`:`payment_preimage`]

For a timed out or route-failed HTLC:

1. type: 131 (`update_fail_htlc`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`u64`:`id`]
   * [`u16`:`len`]
   * [`len*byte`:`reason`]

The `reason` field is an opaque encrypted blob for the benefit of the
original HTLC initiator, as defined in [BOLT #4](04-onion-routing.md);
however, there's a special malformed failure variant for the case where
the peer couldn't parse it: in this case the current node instead takes action, encrypting
it into a `update_fail_htlc` for relaying.

For an unparsable HTLC:

1. type: 135 (`update_fail_malformed_htlc`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`u64`:`id`]
   * [`sha256`:`sha256_of_onion`]
   * [`u16`:`failure_code`]

##### Committing Updates So Far: `commitment_signed`

When a node has changes for the remote commitment, it can apply them,
sign the resulting transaction (as defined in [BOLT #3](03-transactions.md)), and send a
`commitment_signed` message.

1. type: 132 (`commitment_signed`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`signature`:`signature`]
   * [`u16`:`num_htlcs`]
   * [`num_htlcs*signature`:`htlc_signature`]

##### Completing the Transition to the Updated State: `revoke_and_ack`

Once the recipient of `commitment_signed` checks the signature and knows
it has a valid new commitment transaction, it replies with the commitment
preimage for the previous commitment transaction in a `revoke_and_ack`
message.

This message also implicitly serves as an acknowledgment of receipt
of the `commitment_signed`, so this is a logical time for the `commitment_signed` sender
to apply (to its own commitment) any pending updates it sent before
that `commitment_signed`.

The description of key derivation is in [BOLT #3](03-transactions.md#key-derivation).

1. type: 133 (`revoke_and_ack`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`32*byte`:`per_commitment_secret`]
   * [`point`:`next_per_commitment_point`]

##### Updating Fees: `update_fee`

An `update_fee` message is sent by the node which is paying the
Bitcoin fee. Like any update, it's first committed to the receiver's
commitment transaction and then (once acknowledged) committed to the
sender's. Unlike an HTLC, `update_fee` is never closed but simply
replaced.

There is a possibility of a race, as the recipient can add new HTLCs
before it receives the `update_fee`. Under this circumstance, the sender may
not be able to afford the fee on its own commitment transaction, once the `update_fee`
is finally acknowledged by the recipient. In this case, the fee will be less
than the fee rate, as described in [BOLT #3](03-transactions.md#fee-payment).

The exact calculation used for deriving the fee from the fee rate is
given in [BOLT #3](03-transactions.md#fee-calculation).

1. type: 134 (`update_fee`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`u32`:`feerate_per_kw`]

### Message Retransmission

Because communication transports are unreliable, and may need to be
re-established from time to time, the design of the transport has been
explicitly separated from the protocol.

Nonetheless, it's assumed our transport is ordered and reliable.
Reconnection introduces doubt as to what has been received, so there are
explicit acknowledgments at that point.

This is fairly straightforward in the case of channel establishment
and close, where messages have an explicit order, but during normal
operation, acknowledgments of updates are delayed until the
`commitment_signed` / `revoke_and_ack` exchange; so it cannot be assumed
that the updates have been received. This also means that the receiving
node only needs to store updates upon receipt of `commitment_signed`.

Note that messages described in [BOLT #7](07-routing-gossip.md) are
independent of particular channels; their transmission requirements
are covered there, and besides being transmitted after `init` (as all
messages are), they are independent of requirements here.

1. type: 136 (`channel_reestablish`)
2. data:
   * [`channel_id`:`channel_id`]
   * [`u64`:`next_commitment_number`]
   * [`u64`:`next_revocation_number`]
   * [`32*byte`:`your_last_per_commitment_secret`] (option_data_loss_protect)
   * [`point`:`my_current_per_commitment_point`] (option_data_loss_protect)

`next_commitment_number`: A commitment number is a 48-bit
incrementing counter for each commitment transaction; counters
are independent for each peer in the channel and start at 0.
They're only explicitly relayed to the other node in the case of
re-establishment, otherwise they are implicit.


## 3: Bitcoin Transaction and Script Formats

This details the exact format of on-chain transactions, which both sides need to agree on to ensure signatures are valid. This consists of the funding transaction output script, the commitment transactions, and the HTLC transactions.

# Transactions

## Transaction Input and Output Ordering

Lexicographic ordering: see [BIP69](https://github.com/bitcoin/bips/blob/master/bip-0069.mediawiki).  In the case of identical HTLC outputs, the outputs are ordered in increasing `cltv_expiry` order.

## Use of Segwit

Most transaction outputs used here are pay-to-witness-script-hash<sup>[BIP141](https://github.com/bitcoin/bips/blob/master/bip-0141.mediawiki#witness-program)</sup> (P2WSH) outputs: the Segwit version of P2SH. To spend such outputs, the last item on the witness stack must be the actual script that was used to generate the P2WSH output that is being spent. This last item has been omitted for brevity in the rest of this document.

## Funding Transaction Output

* The funding output script is a P2WSH to:

`2 <pubkey1> <pubkey2> 2 OP_CHECKMULTISIG`

* Where `pubkey1` is the numerically lesser of the two DER-encoded `funding_pubkey` and where `pubkey2` is the numerically greater of the two.

## Commitment Transaction

* version: 2
* locktime: upper 8 bits are 0x20, lower 24 bits are the lower 24 bits of the obscured commitment number
* txin count: 1
   * `txin[0]` outpoint: `txid` and `output_index` from `funding_created` message
   * `txin[0]` sequence: upper 8 bits are 0x80, lower 24 bits are upper 24 bits of the obscured commitment number
   * `txin[0]` script bytes: 0
   * `txin[0]` witness: `0 <signature_for_pubkey1> <signature_for_pubkey2>`

The 48-bit commitment number is obscured by `XOR` with the lower 48 bits of:

    SHA256(payment_basepoint from open_channel || payment_basepoint from accept_channel)

This obscures the number of commitments made on the channel in the
case of unilateral close, yet still provides a useful index for both
nodes (who know the `payment_basepoint`s) to quickly find a revoked
commitment transaction.

### Commitment Transaction Outputs

To allow an opportunity for penalty transactions, in case of a revoked commitment transaction, all outputs that return funds to the owner of the commitment transaction (a.k.a. the "local node") must be delayed for `to_self_delay` blocks. This delay is done in a second-stage HTLC transaction (HTLC-success for HTLCs accepted by the local node, HTLC-timeout for HTLCs offered by the local node).

The reason for the separate transaction stage for HTLC outputs is so that HTLCs can timeout or be fulfilled even though they are within the `to_self_delay` delay.
Otherwise, the required minimum timeout on HTLCs is lengthened by this delay, causing longer timeouts for HTLCs traversing the network.

The amounts for each output MUST be rounded down to whole satoshis. If this amount, minus the fees for the HTLC transaction, is less than the `dust_limit_satoshis` set by the owner of the commitment transaction, the output MUST NOT be produced (thus the funds add to fees).

#### `to_local` Output

This output sends funds back to the owner of this commitment transaction and thus must be timelocked using `OP_CHECKSEQUENCEVERIFY`. It can be claimed, without delay, by the other party if they know the revocation private key. The output is a version-0 P2WSH, with a witness script:

    OP_IF
        # Penalty transaction
        <revocationpubkey>
    OP_ELSE
        `to_self_delay`
        OP_CHECKSEQUENCEVERIFY
        OP_DROP
        <local_delayedpubkey>
    OP_ENDIF
    OP_CHECKSIG

The output is spent by a transaction with `nSequence` field set to `to_self_delay` (which can only be valid after that duration has passed) and witness:

    <local_delayedsig> 0

If a revoked commitment transaction is published, the other party can spend this output immediately with the following witness:

    <revocation_sig> 1

#### `to_remote` Output

This output sends funds to the other peer and thus is a simple P2WPKH to `remotepubkey`.

#### Offered HTLC Outputs

This output sends funds to either an HTLC-timeout transaction after the HTLC-timeout or to the remote node using the payment preimage or the revocation key. The output is a P2WSH, with a witness script:

    # To remote node with revocation key
    OP_DUP OP_HASH160 <RIPEMD160(SHA256(revocationpubkey))> OP_EQUAL
    OP_IF
        OP_CHECKSIG
    OP_ELSE
        <remote_htlcpubkey> OP_SWAP OP_SIZE 32 OP_EQUAL
        OP_NOTIF
            # To local node via HTLC-timeout transaction (timelocked).
            OP_DROP 2 OP_SWAP <local_htlcpubkey> 2 OP_CHECKMULTISIG
        OP_ELSE
            # To remote node with preimage.
            OP_HASH160 <RIPEMD160(payment_hash)> OP_EQUALVERIFY
            OP_CHECKSIG
        OP_ENDIF
    OP_ENDIF

The remote node can redeem the HTLC with the witness:

    <remotehtlcsig> <payment_preimage>

If a revoked commitment transaction is published, the remote node can spend this output immediately with the following witness:

    <revocation_sig> <revocationpubkey>

The sending node can use the HTLC-timeout transaction to timeout the HTLC once the HTLC is expired, as shown below.

#### Received HTLC Outputs

This output sends funds to either the remote node after the HTLC-timeout or using the revocation key, or to an HTLC-success transaction with a successful payment preimage. The output is a P2WSH, with a witness script:

    # To remote node with revocation key
    OP_DUP OP_HASH160 <RIPEMD160(SHA256(revocationpubkey))> OP_EQUAL
    OP_IF
        OP_CHECKSIG
    OP_ELSE
        <remote_htlcpubkey> OP_SWAP OP_SIZE 32 OP_EQUAL
        OP_IF
            # To local node via HTLC-success transaction.
            OP_HASH160 <RIPEMD160(payment_hash)> OP_EQUALVERIFY
            2 OP_SWAP <local_htlcpubkey> 2 OP_CHECKMULTISIG
        OP_ELSE
            # To remote node after timeout.
            OP_DROP <cltv_expiry> OP_CHECKLOCKTIMEVERIFY OP_DROP
            OP_CHECKSIG
        OP_ENDIF
    OP_ENDIF

To timeout the HTLC, the remote node spends it with the witness:

    <remotehtlcsig> 0

If a revoked commitment transaction is published, the remote node can spend this output immediately with the following witness:

    <revocation_sig> <revocationpubkey>

To redeem the HTLC, the HTLC-success transaction is used as detailed below.

### Trimmed Outputs

Each peer specifies a `dust_limit_satoshis` below which outputs should
not be produced; these outputs that are not produced are termed "trimmed". A trimmed output is
considered too small to be worth creating and is instead added
to the commitment transaction fee. For HTLCs, it needs to be taken into
account that the second-stage HTLC transaction may also be below the
limit.

## HTLC-Timeout and HTLC-Success Transactions

These HTLC transactions are almost identical, except the HTLC-timeout transaction is timelocked. Both HTLC-timeout/HTLC-success transactions can be spent by a valid penalty transaction.

* version: 2
* locktime: `0` for HTLC-success, `cltv_expiry` for HTLC-timeout
* txin count: 1
   * `txin[0]` outpoint: `txid` of the commitment transaction and `output_index` of the matching HTLC output for the HTLC transaction
   * `txin[0]` sequence: `0`
   * `txin[0]` script bytes: `0`
   * `txin[0]` witness stack: `0 <remotehtlcsig> <localhtlcsig>  <payment_preimage>` for HTLC-success, `0 <remotehtlcsig> <localhtlcsig> 0` for HTLC-timeout
* txout count: 1
   * `txout[0]` amount: the HTLC amount minus fees (see [Fee Calculation](#fee-calculation))
   * `txout[0]` script: version-0 P2WSH with witness script as shown below

The witness script for the output is:

    OP_IF
        # Penalty transaction
        <revocationpubkey>
    OP_ELSE
        `to_self_delay`
        OP_CHECKSEQUENCEVERIFY
        OP_DROP
        <local_delayedpubkey>
    OP_ENDIF
    OP_CHECKSIG

To spend this via penalty, the remote node uses a witness stack `<revocationsig> 1`, and to collect the output, the local node uses an input with nSequence `to_self_delay` and a witness stack `<local_delayedsig> 0`.

## Closing Transaction

Note that there are two possible variants for each node.

* version: 2
* locktime: 0
* txin count: 1
   * `txin[0]` outpoint: `txid` and `output_index` from `funding_created` message
   * `txin[0]` sequence: 0xFFFFFFFF
   * `txin[0]` script bytes: 0
   * `txin[0]` witness: `0 <signature_for_pubkey1> <signature_for_pubkey2>`
* txout count: 0, 1 or 2
   * `txout` amount: final balance to be paid to one node (minus `fee_satoshis` from `closing_signed`, if this peer funded the channel)
   * `txout` script: as specified in that node's `scriptpubkey` in its `shutdown` message

## Fees

### Fee Calculation

The fee calculation for both commitment transactions and HTLC
transactions is based on the current `feerate_per_kw` and the
*expected weight* of the transaction.

The actual and expected weights vary for several reasons:

* Bitcoin uses DER-encoded signatures, which vary in size.
* Bitcoin also uses variable-length integers, so a large number of outputs will take 3 bytes to encode rather than 1.
* The `to_remote` output may be below the dust limit.
* The `to_local` output may be below the dust limit once fees are extracted.

Thus, a simplified formula for *expected weight* is used, which assumes:

* Signatures are 73 bytes long (the maximum length).
* There are a small number of outputs (thus 1 byte to count them).
* There are always both a `to_local` output and a `to_remote` output.

This yields the following *expected weights* (details of the computation in [Appendix A](#appendix-a-expected-weights)):

    Commitment weight:   724 + 172 * num-untrimmed-htlc-outputs
    HTLC-timeout weight: 663
    HTLC-success weight: 703

Note the reference to the "base fee" for a commitment transaction in the requirements below, which is what the funder pays. The actual fee may be higher than the amount calculated here, due to rounding and trimmed outputs.

#### Example

For example, suppose there is a `feerate_per_kw` of 5000, a `dust_limit_satoshis` of 546 satoshis, and a commitment transaction with:
* two offered HTLCs of 5000000 and 1000000 millisatoshis (5000 and 1000 satoshis)
* two received HTLCs of 7000000 and 800000 millisatoshis (7000 and 800 satoshis)

The HTLC-timeout transaction `weight` is 663, and thus the fee is 3315 satoshis.
The HTLC-success transaction `weight` is 703, and thus the fee is 3515 satoshis

The commitment transaction `weight` is calculated as follows:

* `weight` starts at 724.

* The offered HTLC of 5000 satoshis is above 546 + 3315 and results in:
  * an output of 5000 satoshi in the commitment transaction
  * an HTLC-timeout transaction of 5000 - 3315 satoshis that spends this output
  * `weight` increases to 896

* The offered HTLC of 1000 satoshis is below 546 + 3315 so it is trimmed.

* The received HTLC of 7000 satoshis is above 546 + 3515 and results in:
  * an output of 7000 satoshi in the commitment transaction
  * an HTLC-success transaction of 7000 - 3515 satoshis that spends this output
  * `weight` increases to 1068

* The received HTLC of 800 satoshis is below 546 + 3515 so it is trimmed.

The base commitment transaction fee is 5340 satoshi; the actual
fee (which adds the 1000 and 800 satoshi HTLCs that would make dust
outputs) is 7140 satoshi. The final fee may be even higher if the
`to_local` or `to_remote` outputs fall below `dust_limit_satoshis`.

### Fee Payment

Base commitment transaction fees are extracted from the funder's amount; if that amount is insufficient, the entire amount of the funder's output is used.

Note that after the fee amount is subtracted from the to-funder output,
that output may be below `dust_limit_satoshis`, and thus will also
contribute to fees.

A node:
  - if the resulting fee rate is too low:
    - MAY fail the channel.

## Commitment Transaction Construction

This section ties the previous sections together to detail the
algorithm for constructing the commitment transaction for one peer:
given that peer's `dust_limit_satoshis`, the current `feerate_per_kw`,
the amounts due to each peer (`to_local` and `to_remote`), and all
committed HTLCs:

1. Initialize the commitment transaction input and locktime, as specified
   in [Commitment Transaction](#commitment-transaction).
1. Calculate which committed HTLCs need to be trimmed (see [Trimmed Outputs](#trimmed-outputs)).
2. Calculate the base [commitment transaction fee](#fee-calculation).
3. Subtract this base fee from the funder (either `to_local` or `to_remote`),
   with a floor of 0 (see [Fee Payment](#fee-payment)).
3. For every offered HTLC, if it is not trimmed, add an
   [offered HTLC output](#offered-htlc-outputs).
4. For every received HTLC, if it is not trimmed, add an
   [received HTLC output](#received-htlc-outputs).
5. If the `to_local` amount is greater or equal to `dust_limit_satoshis`,
   add a [`to_local` output](#to_local-output).
6. If the `to_remote` amount is greater or equal to `dust_limit_satoshis`,
   add a [`to_remote` output](#to_remote-output).
7. Sort the outputs into [BIP 69+CLTV order](#transaction-input-and-output-ordering).

# Keys

## Key Derivation

Each commitment transaction uses a unique set of keys: `localpubkey` and `remotepubkey`.
The HTLC-success and HTLC-timeout transactions use `local_delayedpubkey` and `revocationpubkey`.
These are changed for every transaction based on the `per_commitment_point`.

The reason for key change is so that trustless watching for revoked
transactions can be outsourced. Such a _watcher_ should not be able to
determine the contents of a commitment transaction — even if the _watcher_ knows
which transaction ID to watch for and can make a reasonable guess
as to which HTLCs and balances may be included. Nonetheless, to
avoid storage of every commitment transaction, a _watcher_ can be given the
`per_commitment_secret` values (which can be stored compactly) and the
`revocation_basepoint` and `delayed_payment_basepoint` used to regenerate
the scripts required for the penalty transaction; thus, a _watcher_ need only be
given (and store) the signatures for each penalty input.

Changing the `localpubkey` and `remotepubkey` every time ensures that commitment
transaction ID cannot be guessed; every commitment transaction uses an ID
in its output script. Splitting the `local_delayedpubkey`, which is required for
the penalty transaction, allows it to be shared with the _watcher_ without
revealing `localpubkey`; even if both peers use the same _watcher_, nothing is revealed.

Finally, even in the case of normal unilateral close, the HTLC-success
and/or HTLC-timeout transactions do not reveal anything to the
_watcher_, as it does not know the corresponding `per_commitment_secret` and
cannot relate the `local_delayedpubkey` or `revocationpubkey` with their bases.

For efficiency, keys are generated from a series of per-commitment secrets
that are generated from a single seed, which allows the receiver to compactly
store them (see [below](#efficient-per-commitment-secret-storage)).

### `localpubkey`, `remotepubkey`, `local_htlcpubkey`, `remote_htlcpubkey`, `local_delayedpubkey`, and `remote_delayedpubkey` Derivation

These pubkeys are simply generated by addition from their base points:

	pubkey = basepoint + SHA256(per_commitment_point || basepoint) * G

The `localpubkey` uses the local node's `payment_basepoint`;
the `remotepubkey` uses the remote node's `payment_basepoint`;
the `local_htlcpubkey` uses the local node's `htlc_basepoint`;
the `remote_htlcpubkey` uses the remote node's `htlc_basepoint`;
the `local_delayedpubkey` uses the local node's `delayed_payment_basepoint`;
and the `remote_delayedpubkey` uses the remote node's `delayed_payment_basepoint`.

The corresponding private keys can be similarly derived, if the basepoint
secrets are known (i.e. the private keys corresponding to `localpubkey`, `local_htlcpubkey`, and `local_delayedpubkey` only):

    privkey = basepoint_secret + SHA256(per_commitment_point || basepoint)

### `revocationpubkey` Derivation

The `revocationpubkey` is a blinded key: when the local node wishes to create a new
commitment for the remote node, it uses its own `revocation_basepoint` and the remote
node's `per_commitment_point` to derive a new `revocationpubkey` for the
commitment. After the remote node reveals the
`per_commitment_secret` used (thereby revoking that commitment), the local node
can then derive the `revocationprivkey`, as it now knows the two secrets
necessary to derive the key (`revocation_basepoint_secret` and
`per_commitment_secret`).

The `per_commitment_point` is generated using elliptic-curve multiplication:

	per_commitment_point = per_commitment_secret * G

And this is used to derive the revocation pubkey from the remote node's
`revocation_basepoint`:

	revocationpubkey = revocation_basepoint * SHA256(revocation_basepoint || per_commitment_point) + per_commitment_point * SHA256(per_commitment_point || revocation_basepoint)

This construction ensures that neither the node providing the
basepoint nor the node providing the `per_commitment_point` can know the
private key without the other node's secret.

The corresponding private key can be derived once the `per_commitment_secret`
is known:

    revocationprivkey = revocation_basepoint_secret * SHA256(revocation_basepoint || per_commitment_point) + per_commitment_secret * SHA256(per_commitment_point || revocation_basepoint)

### Per-commitment Secret Requirements

A node:
  - MUST select an unguessable 256-bit seed for each connection,
  - MUST NOT reveal the seed.

Up to (2^48 - 1) per-commitment secrets can be generated.

The first secret used:
  - MUST be index 281474976710655,
    - and from there, the index is decremented.

The I'th secret P:
  - MUST match the output of this algorithm:
```
generate_from_seed(seed, I):
    P = seed
    for B in 47 down to 0:
        if B set in I:
            flip(B) in P
            P = SHA256(P)
    return P
```

Where "flip(B)" alternates the B'th least significant bit in the value P.

The receiving node:
  - MAY store all previous per-commitment secrets.
  - MAY calculate them from a compact representation, as described below.

## BOLT #5: Recommendations for On-chain Transaction Handling

### Abstract

Lightning allows for two parties (a local node and a remote node) to conduct transactions
off-chain by giving each of the parties a *cross-signed commitment transaction*,
which describes the current state of the channel (basically, the current balance).
This *commitment transaction* is updated every time a new payment is made and
is spendable at all times.

There are three ways a channel can end:

1. The good way (*mutual close*): at some point the local and remote nodes agree
to close the channel. They generate a *closing transaction* (which is similar to a
commitment transaction, but without any pending payments) and publish it on the
blockchain (see [BOLT #2: Channel Close](02-peer-protocol.md#channel-close)).
2. The bad way (*unilateral close*): something goes wrong, possibly without evil
intent on either side. Perhaps one party crashed, for instance. One side
publishes its *latest commitment transaction*.
3. The ugly way (*revoked transaction close*): one of the parties deliberately
tries to cheat, by publishing an *outdated commitment transaction* (presumably,
a prior version, which is more in its favor).

Because Lightning is designed to be trustless, there is no risk of loss of funds
in any of these three cases; provided that the situation is properly handled.
The goal of this document is to explain exactly how a node should react when it
encounters any of the above situations, on-chain.

### General Nomenclature

Any unspent output is considered to be *unresolved* and can be *resolved*
as detailed in this document. Usually this is accomplished by spending it with
another *resolving* transaction. Although, sometimes simply noting the output
for later wallet spending is sufficient, in which case the transaction containing
the output is considered to be its own *resolving* transaction.

Outputs that are *resolved* are considered *irrevocably resolved*
once the remote's *resolving* transaction is included in a block at least 100
deep, on the most-work blockchain. 100 blocks is far greater than the
longest known Bitcoin fork and is the same wait time used for
confirmations of miners' rewards (see [Reference Implementation](https://github.com/bitcoin/bitcoin/blob/4db82b7aab4ad64717f742a7318e3dc6811b41be/src/consensus/tx_verify.cpp#L223)).

### Commitment Transaction

The local and remote nodes each hold a *commitment transaction*. Each of these
commitment transactions has four types of outputs:

1. _local node's main output_: Zero or one output, to pay to the *local node's*
commitment pubkey.
2. _remote node's main output_: Zero or one output, to pay to the *remote node's*
commitment pubkey.
3. _local node's offered HTLCs_: Zero or more pending payments (*HTLCs*), to pay
the *remote node* in return for a payment preimage.
4. _remote node's offered HTLCs_: Zero or more pending payments (*HTLCs*), to
pay the *local node* in return for a payment preimage.

To incentivize the local and remote nodes to cooperate, an `OP_CHECKSEQUENCEVERIFY`
relative timeout encumbers the *local node's outputs* (in the *local node's
commitment transaction*) and the *remote node's outputs* (in the *remote node's
commitment transaction*). So for example, if the local node publishes its
commitment transaction, it will have to wait to claim its own funds,
whereas the remote node will have immediate access to its own funds. As a
consequence, the two commitment transactions are not identical, but they are
(usually) symmetrical.

See [BOLT #3: Commitment Transaction](03-transactions.md#commitment-transaction)
for more details.

### Failing a Channel

Although closing a channel can be accomplished in several ways, the most
efficient is preferred.

Various error cases involve closing a channel. The requirements for sending
error messages to peers are specified in
[BOLT #1: The `error` Message](01-messaging.md#the-error-message).

### Mutual Close Handling

A closing transaction *resolves* the funding transaction output.

In the case of a mutual close, a node need not do anything else, as it has
already agreed to the output, which is sent to its specified `scriptpubkey` (see
[BOLT #2: Closing initiation: `shutdown`](02-peer-protocol.md#closing-initiation-shutdown)).

### Unilateral Close Handling: Local Commitment Transaction

This is the first of two cases involving unilateral closes. In this case, a
node discovers its *local commitment transaction*, which *resolves* the funding
transaction output.

However, a node cannot claim funds from the outputs of a unilateral close that
it initiated, until the `OP_CHECKSEQUENCEVERIFY` delay has passed (as specified
by the remote node's `to_self_delay` field). Where relevant, this situation is
noted below.

#### HTLC Output Handling: Local Commitment, Local Offers

Each HTLC output can only be spent by either the *local offerer*, by using the
HTLC-timeout transaction after it's timed out, or the *remote recipient*, if it
has the payment preimage.

There can be HTLCs which are not represented by any outputs: either
because they were trimmed as dust, or because the transaction has only been
partially committed.

The HTLC output has *timed out* once the depth of the latest block is equal to
or greater than the HTLC `cltv_expiry`.

#### HTLC Output Handling: Local Commitment, Remote Offers

Each HTLC output can only be spent by the recipient, using the HTLC-success
transaction, which it can only populate if it has the payment
preimage. If it doesn't have the preimage (and doesn't discover it), it's
the offerer's responsibility to spend the HTLC output once it's timed out.

There are several possible cases for an offered HTLC:

1. The offerer is NOT irrevocably committed to it. The recipient will usually
   not know the preimage, since it will not forward HTLCs until they're fully
   committed. So using the preimage would reveal that this recipient is the
   final hop; thus, in this case, it's best to allow the HTLC to time out.
2. The offerer is irrevocably committed to the offered HTLC, but the recipient
   has not yet committed to an outgoing HTLC. In this case, the recipient can
   either forward or timeout the offered HTLC.
3. The recipient has committed to an outgoing HTLC, in exchange for the offered
   HTLC. In this case, the recipient must use the preimage, once it receives it
   from the outgoing HTLC; otherwise, it will lose funds by sending an outgoing
   payment without redeeming the incoming payment.

### Unilateral Close Handling: Remote Commitment Transaction

The *remote node's* commitment transaction *resolves* the funding
transaction output.

There are no delays constraining node behavior in this case, so it's simpler for
a node to handle than the case in which it discovers its local commitment
transaction (see [Unilateral Close Handling: Local Commitment Transaction](#unilateral-close-handling-local-commitment-transaction)).

#### HTLC Output Handling: Remote Commitment, Local Offers

Each HTLC output can only be spent by either the *local offerer*, after it's
timed out, or by the *remote recipient*, by using the HTLC-success transaction
if it has the payment preimage.

There can be HTLCs which are not represented by any outputs: either
because the outputs were trimmed as dust, or because the remote node has two
*valid* commitment transactions with differing HTLCs.

The HTLC output has *timed out* once the depth of the latest block is equal to
or greater than the HTLC `cltv_expiry`.

#### HTLC Output Handling: Remote Commitment, Remote Offers

The remote HTLC outputs can only be spent by the local node if it has the
payment preimage. If the local node does not have the preimage (and doesn't
discover it), it's the remote node's responsibility to spend the HTLC output
once it's timed out.

There are actually several possible cases for an offered HTLC:

1. The offerer is not irrevocably committed to it. In this case, the recipient
   usually won't know the preimage, since it won't forward HTLCs until
   they're fully committed. As using the preimage would reveal that
   this recipient is the final hop, it's best to allow the HTLC to time out.
2. The offerer is irrevocably committed to the offered HTLC, but the recipient
   hasn't yet committed to an outgoing HTLC. In this case, the recipient can
   either forward it or wait for it to timeout.
3. The recipient has committed to an outgoing HTLC in exchange for an offered
   HTLC. In this case, the recipient must use the preimage, if it receives it
   from the outgoing HTLC; otherwise, it will lose funds by sending an outgoing
   payment without redeeming the incoming one.

### Revoked Transaction Close Handling

If any node tries to cheat by broadcasting an outdated commitment transaction
(any previous commitment transaction besides the most current one), the other
node in the channel can use its revocation private key to claim all the funds from the
channel's original funding transaction.

#### Penalty Transactions Weight Calculation

There are three different scripts for penalty transactions, with the following
witness weights (details of weight computation are in
[Appendix A](#appendix-a-expected-weights)):

    to_local_penalty_witness: 160 bytes
    offered_htlc_penalty_witness: 243 bytes
    accepted_htlc_penalty_witness: 249 bytes

The penalty *txinput* itself takes up 41 bytes and has a weight of 164 bytes,
which results in the following weights for each input:

    to_local_penalty_input_weight: 324 bytes
    offered_htlc_penalty_input_weight: 407 bytes
    accepted_htlc_penalty_input_weight: 413 bytes

The rest of the penalty transaction takes up 4+1+1+8+1+34+4=53 bytes of
non-witness data: assuming it has a pay-to-witness-script-hash (the largest
standard output script), in addition to a 2-byte witness header.

In addition to spending these outputs, a penalty transaction may optionally
spend the commitment transaction's `to_remote` output (e.g. to reduce the total
amount paid in fees). Doing so requires the inclusion of a P2WPKH witness and an
additional *txinput*, resulting in an additional 108 + 164 = 272 bytes.

In the worst case scenario, the node holds only incoming HTLCs, and the
HTLC-timeout transactions are not published, which forces the node to spend from
the commitment transaction.

With a maximum standard weight of 400000 bytes, the maximum number of HTLCs that
can be swept in a single transaction is as follows:

    max_num_htlcs = (400000 - 324 - 272 - (4 * 53) - 2) / 413 = 966

Thus, 483 bidirectional HTLCs (containing both `to_local` and
`to_remote` outputs) can be resolved in a single penalty transaction.
Note: even if the `to_remote` output is not swept, the resulting
`max_num_htlcs` is 967; which yields the same unidirectional limit of 483 HTLCs.
