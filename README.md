# Fabric Application for Hyperledger Fabric 1.0 which create virtual car object to track build of material from various suppliers.

# Below app can be used to used to build BOm structures for Car,Flight,Shipemnts -Change the orgs,channels and write new chain code to get started


The network can be deployed to multiple docker containers on one host for development or to multiple hosts for testing 
or production.

## Members and Components

Network consortium consists of:

- Orderer organization `example.com`
- Peer organization org1 `dealer` 
- Peer organization org2 `manufacturer` 
- Peer organization org3 `ford`
- Peer organization org2 `gm` 
- Peer organization org3 `rr`(**rolls royce)

They transact with each other on the following channels:

- `common` involving all members and with chaincode `reference` deployed
- bilateral confidential channels between pairs of members with chaincode `relationship` deployed to them
  - `dealer-manufacturer-ford-gm-rr`
  - `manufacturer-ford`
  - `manufacturer-rr`
  - `manufacturer-gm`
