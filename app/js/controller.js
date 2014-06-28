/*jshint jquery: true, browser: true*/

function Timeline($scope, $http) {
	"use strict";

	var pathname = window.location.pathname.split("/");
	var raw_data_url = pathname.slice(0, 4).join("/");

	$scope.colorFunction = function() {
		return function() { return "#7D1935"; };
	};

	var format = d3.format('.2f');
	$scope.yAxisTickFormatFunction = function(){
		return function(d) {
			return format(d);
		};
	};

	$scope.height = window.innerHeight * 0.95;

	$http.get(raw_data_url).success(function(data) {
		var ts0 = 0;
		$scope.max = 0;
		var values = [];

		Object.keys(data).forEach(function (timestamp) {
			var ts = Math.round(parseInt(timestamp) / Math.pow(10, 9));
			if (ts0 === 0) {
				ts0 = ts;
			}
			ts -= ts0;
			$scope.max = Math.max($scope.max, data[timestamp]);
			values.push([ts, data[timestamp]]);
		});

		$scope.max *= 1.1;
		$scope.chartData = [{"key": pathname[3], "values": values}];

		setTimeout(function() {
			$('.nv-point').attr("r", "3.5");
		}, 500);
	});
}
