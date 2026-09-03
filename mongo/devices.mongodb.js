// MongoDB Playground
// Use Ctrl+Space inside a snippet or a string literal to trigger completions.

// The current database to use.
use("blueasset");

db.devices.aggregate([
    {
        $match: {
            $expr: {
                $eq: ['$owner', { '$toObjectId': '69c238c9a84d8ebf1ed509c8' }]
            }
        }
    },
    {
        $project: {
            _id: 1, imei: 1, name: 1, device_type: 1, owner: 1, status: 1, msisdn: 1
        }
    },
    {
        $lookup: {
            from: "profile_devices",
            localField: "owner",
            foreignField: "profile_id",
            as: "not_owned"
        }
    },
    {
        $unwind: {
            path: "$not_owned",
            preserveNullAndEmptyArrays: true
        }
    },
    {
        $project:{
            imei:1
        },
    },
    {
        $lookup: {
            from: "vas_subscriptions",
            let: { owner: "$owner", "imei": "$imei" }, // 1. Pass primary field to variable
            pipeline: [
                {
                    $sort: { "vas": 1 },
                },
                {
                    $match: {
                        $expr: {
                            $and: [
                                { $eq: ["$imei", "$$imei"] },
                                { $eq: ["$profile_id", "$$owner"] }
                            ]
                        }
                    }
                }
            ],
            as: "subscriptions"
        }
    },
    {
        $project: {
            "profile.name": 1, "profile.surname": 1,
            "profile.email": 1, "profile.company": 1,
            'subscriptions.vas': 1,
            owner: 1,
            name: 1,
            msisdn: 1,
            imei: 1,
            device_type: 1
        }
    }])