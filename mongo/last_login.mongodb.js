// MongoDB Playground
// Use Ctrl+Space inside a snippet or a string literal to trigger completions.

// The current database to use.
use("blueasset");

db.profiles.aggregate([
    {
        $match: {
            $expr: {
                $and: [
                    {
                        $eq: [{ $year: "$last_login" }, { $year: "$$NOW" }]
                    },
                    {
                        $eq: [{ $month: "$last_login" }, { $month: "$$NOW" }]
                    }]
            }
        },
    },
    {
        $lookup: {
            from: "devices",
            localField: "_id",
            foreignField: "owner",
            as: "device_info",
        }
    },
    {
        $addFields: {
            device_count: { $size: "$device_info" }
        }
    },
    {
        $project: {
            last_login: 1,
            device_count: 1,
            name: 1,
            surname: 1,
            email: 1,
            mobile_no: 1,
            user_name: 1,

        }
    }

]
)