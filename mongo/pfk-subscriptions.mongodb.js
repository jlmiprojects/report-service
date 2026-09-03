// MongoDB Playground
// Use Ctrl+Space inside a snippet or a string literal to trigger completions.

// The current database to use.
use("blueasset");

db.vas_subscriptions.aggregate([{ $match: { vas: "SVR" } }, { $group: { _id: "$profile_id", vas_count: { $sum: 1 } } }, { $lookup: { from: "profiles", localField: "_id", foreignField: "_id", as: "profile_info" } }, { $unwind: "$profile_info" }, { $project: { _id: 0, user_name: "$profile_info.user_name", name: "$profile_info.name", surname: "$profile_info.surname", vas_count: 1 } }])
