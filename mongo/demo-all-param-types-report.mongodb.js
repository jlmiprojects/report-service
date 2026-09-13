// Demo report exercising every supported report-parameter type in one form:
// TEXT, NUMBER, DATE, BOOL, SELECT, RADIO, CHECKBOX, LOOKUP. Open it from the
// broker-portal Reports page ("Run") to see each control render, and watch
// "brokerage_id" (LOOKUP) live-refresh its options whenever "country"
// (SELECT) changes — the cascading wired up alongside the LOOKUP endpoint.
// Run once from the VS Code MongoDB extension (or paste into mongosh), then
// remove with the db.reports.deleteOne(...) at the bottom when done.
//
// Prereqs: conf/application.json -> mongo_connections.broker_portal points at
// the cluster hosting the broker-portal `broker_portal` database (the
// "brokerages" lookup script queries its `brokerages` collection).

use("blueasset");

db.reports.updateOne(
    { name: "demo-all-param-types" },
    {
        $set: {
            name: "demo-all-param-types",
            handler: "generic",
            description: "Demo report exercising every supported parameter type: TEXT, NUMBER, DATE, BOOL, SELECT, RADIO, CHECKBOX, LOOKUP.",
            template: "demo_all_param_types",
            script: "demo_all_param_types",
            management: false,
            excel: false,
            parameters: [
                {
                    name: "full_name",
                    description: "TEXT — letters and spaces only.",
                    required: true,
                    type: "text",
                    regexp: "^[A-Za-z ]{2,50}$"
                },
                {
                    name: "age",
                    description: "NUMBER.",
                    required: false,
                    type: "number"
                },
                {
                    name: "start_date",
                    description: "DATE.",
                    required: false,
                    type: "date"
                },
                {
                    name: "is_urgent",
                    description: "BOOL — a single yes/no checkbox.",
                    required: false,
                    type: "bool"
                },
                {
                    name: "country",
                    description: "SELECT — static dropdown. Changing this live-rescopes the brokerage LOOKUP below.",
                    required: false,
                    type: "select",
                    metadata: {
                        options: [
                            { name: "South Africa", value: "ZA" },
                            { name: "Namibia", value: "NA" },
                            { name: "Botswana", value: "BW" }
                        ]
                    }
                },
                {
                    name: "priority",
                    description: "RADIO — static radio-button group.",
                    required: false,
                    type: "radio",
                    metadata: {
                        options: [
                            { name: "Low", value: "low" },
                            { name: "Medium", value: "medium" },
                            { name: "High", value: "high" }
                        ]
                    }
                },
                {
                    name: "tags",
                    description: "CHECKBOX — static multi-select checkbox group.",
                    required: false,
                    type: "checkbox",
                    metadata: {
                        options: [
                            { name: "Reporting", value: "reporting" },
                            { name: "Compliance", value: "compliance" },
                            { name: "Finance", value: "finance" },
                            { name: "Ops", value: "ops" }
                        ]
                    }
                },
                {
                    name: "brokerage_id",
                    description: "LOOKUP — resolved server-side by the 'brokerages' yaegi script (scripts/brokerages.go), re-resolved live whenever 'country' changes.",
                    required: false,
                    type: "lookup",
                    metadata: {
                        script: "brokerages",
                        display: "select"
                    }
                }
            ]
        }
    },
    { upsert: true }
);

db.reports.findOne({ name: "demo-all-param-types" });

// When done demoing:
// db.reports.deleteOne({ name: "demo-all-param-types" });
