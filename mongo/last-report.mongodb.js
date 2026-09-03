use("blueasset");


db.last.aggregate([
    {
        "$lookup":
            {   "from":"devices",
                "localField":"imei",
                "foreignField":"imei",
                "as":"device"}
            },
    {
        "$unwind":{
            "path":"$device"
        }
    },
    {
        "$project":
            {
                "device.name":1,
                "device.status":1,
                "imei":1,
                "device.comment":1,
                "device.description":1,
                "timestamp":1,
            }
        }])