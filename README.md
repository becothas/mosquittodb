# Go MosquittoDB 

This project aims to create a library for reading and writing mosquitto databases and
perhaps provide some documentation about the mosquitto db internals.

- [x] Reading (partially)
- [ ] Writing

## Database file layout

The database always begins with a DBHeader, following with chunk headers and their data
in between.

|File Layout    |
|:-------------:|
|DbHeader       |
|Chunk Header0  |
|Chunk Data0    |
|Chunk Header1  |
|Chunk Data1    |
|Chunk HeaderN  |
|Chunk DataN    |
|.....          |

The database file layout looks like it has been stable and not changed anything over versions.
Most differences are slight changes to how chunks looks like.

### Chunk IDs:
1. Configuration Chunk
2. Message Store
3. Client Message
4. Retain
5. Subscription
6. Client

### DBHeader

This piece contains a database file magic, a CRC and a Version.
These fields should always be first and always be present in a mosquitto database.

|Name|Type|Notes|
|----|----|-----|
|Magic|[15]byte|Always "\x00\xB5\x00mosquitto db"|
|CRC|uint32|I don't see this crc used anywhere|
|Version|uint32| Version of database. At time of writing 6 is the latest|

### ChunkHeader

This header contains a chunk type id, and the chunk length.
Having the chunk length enables skipping unknown / unwanted chunks.


#### V5, V6
|Name|Type|Notes|
|----|----|-----|
|Type|uint32|One of the Chunk type IDs|
|Length|uint32|Length of the chunk|

#### V2, V3, V4
|Name|Type|Notes|
|----|----|-----|
|Type|uint16|One of the Chunk type IDs|
|Length|uint32|Length of the chunk|

### Config Chunk

I'm guessing this chunk always comes right after the DBHeader, and only one of this
chunk is ever present in the file.

This chunk contains information about last allocated StoreID, whether the broker shut down 
properly and the size of a StoreID.

#### V5, V6
|Name|Type|Notes|
|----|----|-----|
|LastStoreID|uint64|Varies in size?|
|Shutdown|uint8|this field is 1 if mosquitto was shut down properly|
|StoreIDSize|uint8|Size of a StoreID|

#### V2, V3, V4
|Name|Type|Notes|
|----|----|-----|
|Shutdown|uint8|this field is 1 if mosquitto was shut down properly|
|StoreIDSize|uint8|Size of a StoreID|
|LastStoreID|uint64|Varies in size?|

### Message Store Chunk

This chunk contains one stored message.

#### V5, V6
|Name|Type|Notes|
|----|----|-----|
|StoreID|uint64|Message store ID|
|ExpiryTime|int64|Timestamp when this message is expired|
|PayloadLen|uint32|Payload length|
|SourceMid|uint16|Source message id|
|SourceIDLen|uint16|Source client id length|
|SourceUsernameLen|uint16|Source username length|
|TopicLen|uint16|Topic length|
|SourcePort|uint16|Source port|
|QoS|uint8|Quality of service|
|Retain|uint8|is 1 if this message is retained|
|SourceID|string|contains the source client id. size comes from `SourceIDLen` field|
|SourceUsername|string| contains the source username. size comes from `SourceUsernameLen`|
|Topic|string|Message topic. size comes from `TopicLen` field|
|Payload|[]byte|payload data. size comes from `PayloadLen` field|
|Properties|-|this field covers the rest of the chunk length. This version of the library skips this field|

#### V4
|Name|Type|Notes|
|----|----|-----|
|StoreID|uint64|Message store ID|
|SourceIDLen|uint16|Source client id length|
|SourceID|string|contains the source client id. size comes from `SourceIDLen` field|
|SourceUsernameLen|uint16|**V4 only!** Source username length|
|SourceUsername|string|**V4 only!** Source username. size comes from `SourceUsernameLen` field|
|SourcePort|uint16|**V4 only!** Source port|
|SourceMid|uint16|Source message id|
|MID|uint16|This field is never used and is not stored anywhere when read|
|TopicLen|uint16|Topic length|
|Topic|string|Message topic. size comes from `TopicLen` field|
|QoS|uint8|Quality of service|
|Retain|uint8|is 1 if this message is retained|
|PayloadLen|uint32|Payload length|
|Payload|[]byte|payload data. size comes from `PayloadLen` field|

#### V2, V3
|Name|Type|Notes|
|----|----|-----|
|StoreID|uint64|Message store ID|
|SourceIDLen|uint16|Source client id length|
|SourceID|string|contains the source client id. size comes from `SourceIDLen` field|
|SourceMid|uint16|Source message id|
|MID|uint16|This field is never used and is not stored anywhere when read|
|TopicLen|uint16|Topic length|
|Topic|string|Message topic. size comes from `TopicLen` field|
|QoS|uint8|Quality of service|
|Retain|uint8|is 1 if this message is retained|
|PayloadLen|uint32|Payload length|
|Payload|[]byte|payload data. size comes from `PayloadLen` field|

### Client Message Chunk

#### V5, V6
|Name|Type|Notes|
|----|----|-----|
|StoreID|uint64|Message store ID|
|MID|uint16|Message ID|
|IDLen|uint16|Client id length|
|QoS|uint8|Quality of Service|
|State|uint8|Some internal message state for mosquitto. TODO: enumerate|
|RetainDup|uint8|Retention: (flags & 0xF0 >> 4) Dup: (flags & 0x0F)|
|Direction|uint8|Direction of message. TODO: enumerate|
|ClientID|string|ClientID for message. size comes from `IDLen` field|

#### V2, V3, V4
|Name|Type|Notes|
|----|----|-----|
|IDLen|uint16|Client id length|
|ClientID|string|ClientID for message. size comes from `IDLen` field|
|StoreID|uint64|Message store ID|
|MID|uint16|Message ID|
|QoS|uint8|Quality of Service|
|Retain|uint8|Retain flag|
|Direction|uint8|Direction of message. TODO: enumerate|
|State|uint8|Some internal message state for mosquitto. TODO: enumerate|
|Dup|uint8|Duplicate flag|



### Retain Chunk 

This chunk contains an ID for a stored message, that is retained.

I'm unsure why this chunk is necessary. 
Both the Stored Message chunk, and the Client message chunk contains a retain-flag.

The StoreID seems to refer to a Message in a store chunk.

#### V2,V3, V4, V5, V6

|Name|Type|Notes|
|----|----|-----|
|StoreID|uint64|Message store ID|


### Subscription Chunk

#### V5, V6

|Name|Type|Notes|
|----|----|-----|
|Identifier|uint32|??Not sure|
|IDLen|uint16|Length of client id|
|TopicLen|uint16|Length of topic|
|QoS|uint8|Quality of service|
|Options|uint8|Subscription options|
|ClientID|string|Client id for subscription. size comes from `IDLen` field|
|Topic|string|Subscription topic. size comes from `TopicLen` field|


#### V2, V3, V4
|Name|Type|Notes|
|----|----|-----|
|IDLen|uint16|Length of client id|
|ClientID|string|Client id for subscription. size comes from `IDLen` field|
|TopicLen|uint16|Length of topic|
|Topic|string|Subscription topic. size comes from `TopicLen` field|
|QoS|uint8|Quality of service|

### Client Chunk

This chunk keeps information about one client.

#### V6

|Name|Type|Notes|
|----|----|-----|
|SessionExpiryTime|int64|Timestamp when the client is considered expired|
|SessionExpiryInterval|uint32|??Not sure what this is|
|LastMID|uint16|Last MessageID for client|
|IDLen|uint16|Length of client id|
|ListenerPort|uint16|Listener port. **new in v6**|
|UsernameLen|uint16|Username length. **new in v6**|
|Padding|uint32|Quote from mosquitto source: "tail: 4 byte padding, because 64bit member forces multiple of 8 for struct size"|
|ClientID|string|Client id. size comes from `IDLen` field|
|Username|string|Username of client. size comes from `UsernameLen` field. **new in v6**|

#### V5

|Name|Type|Notes|
|----|----|-----|
|SessionExpiryTime|int64|Timestamp when the client is considered expired|
|SessionExpiryInterval|uint32|??Not sure what this is|
|LastMID|uint16|Last MessageID for client|
|IDLen|uint16|Length of client id|
|ClientID|string|Client id. size comes from `IDLen` field|

#### V3, V4
|Name|Type|Notes|
|----|----|-----|
|IDLen|uint16|Length of client id|
|ClientID|string|Client id. size comes from `IDLen` field|
|Time|**time_t**|Presumably the time this message was sent. This field appears not to be saved anywhere and just read then ignored.|

#### V2 
|Name|Type|Notes|
|----|----|-----|
|IDLen|uint16|Length of client id|
|ClientID|string|Client id. size comes from `IDLen` field|
