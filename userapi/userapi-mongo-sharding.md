## MongoDB sharding for User signup API :

A MongoDB sharded cluster consists of the following components:

1. shards:  Each shard contains a subset of the sharded data. Each shard is deployed as a replica set. <br/>
2. mongos: The mongos acts as a query router, providing an interface between client applications and the sharded cluster.<br/>
3. config servers: Config servers store metadata and configuration settings for the cluster. As of MongoDB 3.4, config servers must be deployed as a replica set (CSRS). <br/>

Reference : https://docs.mongodb.com/manual/sharding/

## Below are 5 instances for MongoDB sharding:
```
	config-server1                       : 10.0.1.93
	config-server2                       : 10.0.1.17
	shard-Replica set 1 (Primary)	     : 10.0.1.194
	shard-Replica set 1 (Secondary)	     : 10.0.1.46
	shard-Replica set 2 (Primary)	     : 10.0.1.178
	shard-Replica set 2 (Secondary)	     : 10.0.1.241
	mongos	                             : 34.212.50.122
```
## AWS console screen:
![](https://github.com/nguyensjsu/sp19-281-mavericks/blob/master/images/mongo-shard-instances-aws.PNG)

## Sharding status:
```
	sh.status()
```
![](https://github.com/nguyensjsu/sp19-281-mavericks/blob/master/images/user-mongo-shard.png)

## List Shards:
```
	db.adminCommand({listShards:1})
```
![](https://github.com/nguyensjsu/sp19-281-mavericks/blob/master/images/user-mongo-list-shards.png)

## Shard distribution:
```
	db.users.getShardDistribution()
```
![](https://github.com/nguyensjsu/sp19-281-mavericks/blob/master/images/user-mongo-shard-distribution.png)

## Records in shard replica set 1
![](https://github.com/nguyensjsu/sp19-281-mavericks/blob/master/images/user-mongo-shard1-records.png)

## Records in shard replica set 2
![](https://github.com/nguyensjsu/sp19-281-mavericks/blob/master/images/user-mongo-shard2-records.png)
