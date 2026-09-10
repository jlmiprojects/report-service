// Registers (upserts) the "record-of-advice" report definition in the
// blueasset.reports collection. Run once from the VS Code MongoDB extension
// (or paste into mongosh). Mirrors conf/record_of_advice_report.json.
//
// Prereqs: conf/application.json -> mongo_connections.broker_portal points at
// the cluster hosting the broker-portal `broker_portal` database, and
// reports/record_of_advice.html + scripts/record_of_advice.js are deployed.
// Restart reports-service after adding the template (parsed once at boot).

use("blueasset");

db.reports.updateOne(
    { name: "record-of-advice" },
    {
        $set: {
            name: "record-of-advice",
            handler: "generic",
            description: "FAIS Record of Advice for a completed Financial Needs Analysis.",
            template: "record_of_advice",
            management: false,
            parameters: [
                {
                    name: "fna_id",
                    description: "The FNA _id (hex ObjectID) to produce the Record of Advice for.",
                    required: true,
                    type: "string"
                }
            ],
            data_actions: [
                {
                    name: "fna",
                    type: "mongo",
                    action: "mongo.aggregate",
                    connection: "broker_portal",
                    ttl: "20s",
                    request: {
                        collection: "fna",
                        pipeline: [
                            { $match: { $expr: { $eq: ["$_id", { $toObjectId: "{{.fna_id}}" }] } } },
                            { $limit: 1 },
                            { $lookup: { from: "clients", localField: "client_id", foreignField: "_id", as: "client" } },
                            { $unwind: { path: "$client", preserveNullAndEmptyArrays: true } },
                            { $lookup: { from: "advisors", localField: "advisor_id", foreignField: "_id", as: "advisor" } },
                            { $unwind: { path: "$advisor", preserveNullAndEmptyArrays: true } },
                            { $lookup: { from: "brokerages", localField: "advisor.brokerage_id", foreignField: "_id", as: "brokerage" } },
                            { $unwind: { path: "$brokerage", preserveNullAndEmptyArrays: true } },
                            {
                                $lookup: {
                                    from: "quotes",
                                    let: { fid: "$_id" },
                                    pipeline: [
                                        { $match: { $expr: { $and: [{ $eq: ["$fna_id", "$$fid"] }, { $eq: ["$status", "SUBMITTED"] }] } } },
                                        { $sort: { created_at: -1 } },
                                        { $limit: 1 }
                                    ],
                                    as: "quote"
                                }
                            },
                            { $unwind: { path: "$quote", preserveNullAndEmptyArrays: true } }
                        ]
                    }
                },
                { name: "record_of_advice", type: "js", action: "record_of_advice" }
            ]
        }
    },
    { upsert: true }
);

db.reports.findOne({ name: "record-of-advice" });
