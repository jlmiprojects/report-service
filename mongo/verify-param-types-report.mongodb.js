// Throwaway report for manually verifying the new report-parameter types
// (LOOKUP / RADIO / CHECKBOX / regexp) end to end. Not a real report — no
// template, no data_actions of its own consequence — just enough to open the
// broker-portal "Run report" dialog and inspect every new/changed control.
// Run once from the VS Code MongoDB extension (or paste into mongosh), then
// remove with the db.reports.deleteOne(...) at the bottom when done.
//
// Prereqs: conf/application.json -> mongo_connections.broker_portal points at
// the cluster hosting the broker-portal `broker_portal` database (same one
// clients_report.json / record_of_advice_report.json use), since the
// "brokerage" lookup below queries its `brokerages` collection.

use("blueasset");

db.reports.updateOne(
    { name: "verify-param-types" },
    {
        $set: {
            name: "verify-param-types",
            handler: "generic",
            description: "Throwaway report for manually verifying LOOKUP/RADIO/CHECKBOX/regexp parameter rendering.",
            template: "",
            management: false,
            excel: false,
            parameters: [
                {
                    name: "reference",
                    description: "Format: three letters, dash, four digits.",
                    required: true,
                    type: "text",
                    regexp: "^[A-Z]{3}-[0-9]{4}$"
                },
                {
                    name: "country",
                    description: "Static RADIO group.",
                    required: false,
                    type: "radio",
                    metadata: {
                        options: [
                            { name: "South Africa", value: "ZA" },
                            { name: "Namibia", value: "NA" },
                            { name: "Botswana", value: "BW" }
                        ]
                    }
                },
                {
                    name: "sections",
                    description: "Static CHECKBOX group (multi-select).",
                    required: false,
                    type: "checkbox",
                    metadata: {
                        options: [
                            { name: "Summary", value: "summary" },
                            { name: "Commission lines", value: "commission" },
                            { name: "Risk breakdown", value: "risk" }
                        ]
                    }
                },
                {
                    name: "brokerage_id",
                    description: "LOOKUP: resolved server-side by the 'brokerages' yaegi script (scripts/brokerages.go), re-resolved live whenever 'country' changes.",
                    required: false,
                    type: "lookup",
                    metadata: {
                        script: "brokerages",
                        display: "select"
                    }
                }
            ],
            data_actions: []
        }
    },
    { upsert: true }
);

db.reports.findOne({ name: "verify-param-types" });

// When done verifying:
// db.reports.deleteOne({ name: "verify-param-types" });
