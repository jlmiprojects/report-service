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
        $unionWith: {
            coll: "profile_devices", pipeline: [
                {
                    $match: {
                        $expr: {
                            $eq: ['$profile_id', { '$toObjectId': '69c238c9a84d8ebf1ed509c8' }]
                        }
                    }
                },
                { $lookup: { from: 'devices', localField: 'imei', foreignField: 'imei', as: 'profile_devices' } },
                { $unwind: { path: "$profile_devices", preserveNullAndEmptyArrays: true } },
                { $replaceRoot: { newRoot: { $mergeObjects: ["$$ROOT", "$profile_devices"] } } },
                { $project: { "profile_devices": 0, "imeis": 0 } }
            ]
        }
    },
    {
        $lookup: {
            from: "vas_subscriptions",
            let: {
                sub_profile: "$owner",
                sub_imei: "$imei"
            },
            pipeline: [
                {
                    $match: {
                        $expr: {
                            $and: [
                                { $eq: ["$imei", "$$sub_imei"] },
                                { $eq: ["$profile_id", "$$sub_profile"] }
                            ]
                        }
                    },

                }
            ],
            as: "subscriptions"
        }
    },

    {
        $addFields: {
            subscriptions: {
                $map: {
                    input: "$subscriptions",
                    as: "sub",
                    in: "$$sub.vas"
                }
            }
        }
    },
    {
        $lookup: {
            from: 'profiles', localField: 'owner',
            foreignField: '_id', as: 'owner_profile',
            pipeline: [
                {
                    $project: { name: 1, surname: 1, _id: 0, "email": 1, "mobile_no": 1 }
                },

            ]
        }
    },
    { $unwind: { path: "$owner_profile", preserveNullAndEmptyArrays: true } },

    {
        $set: {
            "mydriver": {
                "$cond": {
                    "if": { "$eq": ["$driver", ""] },
                    "then": "$owner_profile",
                    "else": "$driver"
                }
            }
        }
    },



    {
        $project: {
            'subscriptions._id': 0, 'subscriptions.imei': 0, 'subscriptions.imeis': 0,
            'subscriptions.settings': 0, 'subscriptions.profile_id': 0,
            'subscriptions.last_updated': 0, 'subscriptions.alerts': 0
        }
    }/*,
    {
              $project: {
                'subscriptions.vas': 1,
                profile_id:1,
                name: 1,
                imei: 1,
                msisdn: 1,
                device_type: 1
              }
    }*/])