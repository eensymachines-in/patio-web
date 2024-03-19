(function () {
    angular.module("patio-app").controller("loginCtrl", function ($scope, $http, $location) {
        $scope.loginErr = null
        $scope.validationErr = false
        $scope.login = {
            usrid: "",
            passwd: "",
            submit: function () {
                // send the login creds back to where this page came from and get the result
                // console.log(this.usrid)
                // console.log(this.passwd)

                if (this.usrid == "" || this.passwd == "") {
                    $scope.validationErr = true;
                    return
                }
                $http({
                    method: 'post',
                    url: '/api/login',
                    data: { u: this.usrid, p: this.passwd },
                    headers: {
                        'Content-Type': "application/json"
                    },
                }).then(function (response) {
                    $location.url("/settings") // logging in to the settings page 
                }, function (response) {
                    if (response.status == 401) {
                        // when the credentials dont match or credentials dont exists 
                        $scope.loginErr = {
                            msg: "Either your credentials did not match or you arent registered on the system"
                        }
                    } else if (response.status == 500) {
                        $scope.loginErr = {
                            msg: "One or more things on the server failed, couldnt sign in"
                        }
                    } else {
                        $scope.loginErr = {
                            msg: "Something went wrong on the server, and we have no idea what!"
                        }
                    }
                })
            }
        }
    })
        .controller("settingsCtrl", function ($scope, $http, $timeout, $route) {
            $scope.hrOptions = [
                "00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12",
                "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"
            ];
            $scope.mnOptions = [
                "00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12",
                "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25",
                "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38",
                "39", "40", "41", "42", "43", "45", "46", "47", "48", "49", "50", "51", "52",
                "53", "54", "55", "56", "57", "58", "59"
            ]
            $scope.configOptions = [
                { opt: 0, txt: "Tick Every Interval", note: "One trigger after every interval, time of the day is irrelevant.", exmp:"Example: ON for 50 seconds, followed by OFF for 50 seconds (interval) in an infinite cycle." },
                { opt: 1, txt: "Tick Every Day At", note: "One trigger at specific time of day, interval is irrelevant.", exmp:"Example: ON at 10:00 one day, OFF the next day at 10:00 in an infinite cycle." },
                { opt: 2, txt: "Pulse Every Interval", note: "Two triggers separated by pulse gap after every interval.", exmp:"Example: ON for 50 seconds (pulse), OFF for 30 seconds, repeat after every 80 seconds (interval) in an infinite cycle." },
                { opt: 3, txt: "Pulse Every Day At", note: "Two triggers separated by pulse gap at specific time of the day", exmp:"Example: ON for 50 seconds, (pulse) OFF after that every day at 10:00, repeat for each day." }
            ]
            $scope.viewModel = { // the model that we used to communicate to the view 
                // viewModel is populated with dedfault values, but when the settings are downloaded this will actual current values
                clock: null,
                CfgOpt: $scope.configOptions[0],
                pulsegap: 60,
                interval: 100,
                // for gui purpose 
                pulsegapInvalid: false,
                intervalInvalid: false
            }
            $scope.settings = { //the model that shall be dispatched away from the controller mostly to the server
                // pump settings 
                config: 0,
                tickat: "",
                pulsegap: 60,
                interval: 100,
            }
            $scope.$watch("viewModel.pulsegap", function (after, before) {
                if (after <= 0 || after > 86340) {
                    $scope.viewModel.pulsegapInvalid = true;
                    return
                }
                // cannot have a pulse gap which is 0 
                if ($scope.viewModel.CfgOpt.opt == 2) {
                    if (after >= $scope.viewModel.interval) {
                        $scope.viewModel.pulsegapInvalid = true;
                        return
                    }
                }
                $scope.viewModel.pulsegapInvalid = false;
                $scope.settings.pulsegap = after
            })
            $scope.$watch("viewModel.interval", function (after, before) {
                if (after <= 0 || after > 86400) {
                    console.log("out of bounds interval")
                    $scope.viewModel.intervalInvalid = true;
                    return
                }
                if ($scope.viewModel.CfgOpt.opt == 2) {
                    if (after <= $scope.viewModel.pulsegap) {
                        console.log("interval cannot be equal to pulse gap")
                        $scope.viewModel.intervalInvalid = true;
                        return
                    }
                }
                $scope.viewModel.intervalInvalid = false;
                $scope.settings.interval = after
            })
            $scope.$watch("viewModel.CfgOpt", function (after, before) {
                if (after !== undefined && after !== null) {
                    $scope.settings.config = after.opt;
                }
            })
            $scope.$watch("viewModel.clock", function (after, before) {
                if (after !== undefined && after !== null) {
                    $scope.settings.tickat = after.hr + ":" + after.min;
                }
            }, true) // its a deep watch since we want to track the properties
            $scope.submit = function () {
                console.log($scope.settings);
                $http({
                    method: 'put',
                    url: '/api/devices/5646564dfgdf/config',
                    data: $scope.settings,
                    headers: {
                        'Content-Type': "application/json"
                    },
                }).then(function(response){
                    console.log("done! settings have been updated");
                    $route.reload(); // reload the same settings page 
                }, function(response){
                    console.log("failed ! settings could not be updated");
                })
            }
            // Getting the current settings to start with 
            $http({
                method: 'get',
                url: '/api/devices/5646564dfgdf/config',
            }).then(function (response) {
                console.log("received current settings from the server", response.data);
                $scope.configOptions.forEach(e => {
                    if (e.opt == response.data.config) {
                        console.log("found matching config option..")
                        $scope.viewModel.CfgOpt = e;
                        return
                    }
                })
                $scope.viewModel.clock= {hr: response.data.tickat.split(":")[0], min :response.data.tickat.split(":")[1]};
                $scope.viewModel.pulsegap = response.data.pulsegap;
                $scope.viewModel.interval  = response.data.interval;

                $timeout(function(){
                    console.log("viewmodel to settings ...");
                    console.log( $scope.settings);
                }, 500)
            }, function (response) {
                console.error("failed to get settings from the device..")
                console.log(response.status)
            })
            
        })
})();