use("blueasset");


db.alerts.aggregate(
    [
        {
            $match:{
                "alert_event.status":
                {
                    $ne:'CLOSED'
                }
            }
        },
        {
            $lookup: {
                from: "profiles",
                localField: "profile_id",
                foreignField: "_id",
                as: "profile_info"
            }
        },
        {
        $unwind: "$profile_info"
        },
        {
            $lookup: {
                from: "devices",
                localField: "imei",
                foreignField: "imei",
                as: "device_info"
            }
        },
        {
        $unwind: "$device_info"
        },
       {
            $project:{
                date_created: "$timestamp",
                alert_type:"$alert_type",
                name:"$profile_info.name",
                surname:"$profile_info.surname",
                device_name:"$device_info.name"
            }
        }
])