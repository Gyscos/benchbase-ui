// Prepare query
getBenchList();

function getBenchList(host) {
	var filter = ""
	$.getJSON(host + "/list?filter=" + filter, function(data) {
		if (data["Error"] != null) {
			// Error?
			alert(data["Error"]);
			return;
		}

		// r is now a list of benchmarks
		var r = data["Result"];

		// Group then by compatible configuration?
	});
}

function groupByConfiguration(benchlist) {

	var result = [];

	var map = {"true":[], "false":[]};

	$.each(benchlist, function(i, data) {
		map[data["Conf"]["ForceAnalyze"]].push(data);
	});


}
