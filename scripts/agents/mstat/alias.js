db_ = db.getSiblingDB("midori")
c = db_.getCollection("alias")
alias = [
{"_id":"i", "v":"Inserts"},
{"_id":"q", "v":"Queries"},
{"_id":"u", "v":"Updates"},
{"_id":"d", "v":"Deletes"},
{"_id":"g", "v": "Getmore ops"},
{"_id":"c", "v":"Command"},
{"_id":"f", "v":"Flushes"},
{"_id":"m", "v":"Mapped"},
{"_id":"v", "v":"Virtual Memory"},
{"_id":"r", "v":"Resident Memory"},
{"_id":"pf", "v":"Page Faults"},
{"_id":"lp", "v":"Locked Percent"},
{"_id":"im", "v":"Ids Miss"},
{"_id":"rq", "v":"Read Queue"},
{"_id":"wq", "v":"Write Queue"},
{"_id":"ar", "v":"Active Read Clients"},
{"_id":"aw", "v":"Active Write Clients"},
{"_id":"ni", "v":"Traffic In (bytes)"},
{"_id":"no", "v":"Traffic Out (bytes)"},
{"_id":"cn", "v":"Open Connections"},
{"_id":"s", "v":"Replica Set Name"},
{"_id":"repl", "v":"Replication Status"}
]
for (var i = 0;i < alias.length;i++){
    c.update({"_id":alias[i]["_id"]}, alias[i], {"upsert":true, "multi":false})
}
