//c/\onsole.log(data.history.trips[0])
console.log("Starting",data.history.trips.length)

data.groups = {}


for (let i = 0; i < data.history.trips.length; i++) {

    let ds = data.history.trips[i].start.substr(0,10);

    if (data.groups[ds] != undefined) {

        data.groups[ds].distance += data.history.trips[i].distance
        data.groups[ds].odo_end = data.history.trips[i].odo_end

        if (data.history.trips[i].business) {
            data.groups[ds].business += data.history.trips[i].distance
        } else {
            data.groups[ds].private += data.history.trips[i].distance
        }

         data.groups[ds].working_time +=  data.history.trips[i].working_time


    } else {
        data.groups[ds] = {"date":ds} //
        data.groups[ds].distance = data.history.trips[i].distance
        data.groups[ds].odo_start = data.history.trips[i].odo_start
        data.groups[ds].odo_end = data.history.trips[i].odo_end
        data.groups[ds].private = 0.0
        data.groups[ds].business  = 0.0
         data.groups[ds].working_time =  data.history.trips[i].working_time

        if (data.history.trips[i].business) {
            data.groups[ds].business = data.history.trips[i].distance
        } else {
            data.groups[ds].private = data.history.trips[i].distance
        }

    }
}

data.history.trips = {}

