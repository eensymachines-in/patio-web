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

                if (this.usrid == "" || this.passwd == ""){
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
                    if (response.status  == 401) {
                        // when the credentials dont match or credentials dont exists 
                        $scope.loginErr = {
                            msg : "Either your credentials did not match or you arent registered on the system"
                        }
                    } else if (response.status == 500) {
                        $scope.loginErr = {
                            msg : "One or more things on the server failed, couldnt sign in"
                        }
                    } else {
                        $scope.loginErr = {
                            msg : "Something went wrong on the server, and we have no idea what!"
                        }
                    }
                })
            }
        }
    })
})();