define(function () {
	app.registerController('UsersCtrl', ['$scope', '$http', '$uibModal', '$rootScope', 'SweetAlert', function ($scope, $http, $modal, $rootScope, SweetAlert) {
		$http.get('/users').then(function (response) {
			$scope.users = response.users;
		});

		$scope.addUser = function () {
			var scope = $rootScope.$new();
			scope.user = {}

			$modal.open({
				templateUrl: '/tpl/users/add.html',
				scope: scope
			}).result.then(function (_response) {
			  var _user = _response.data;
				$http.post('/users', _user).then(function (response) {
					$scope.users.push(response.user);

					$http.post('/users/' + response.user.id + '/password', {
						password: _user.password
					}).catch(function (errorResponse) {
						SweetAlert.swal('Error', 'Setting password failed, API responded with HTTP ' + errorResponse.status, 'error');
					});
				}).error(function (response) {
					SweetAlert.swal('Error', 'API responded with HTTP ' + response.status, 'error');
				});
			}, function () {});
		}
	}]);
});
