var socket = null;
var userId = null;
var rematchSent = false;
var gameOver = false;
var pieceColor = {
	0: "red",
	1: "black",
};

function reset_board() {
	$('#game').empty();
	$('#game').append('<div class="row"><div id="sidebar" class="col-2"><span id="turn_label"></p></div><div class="col-10"><table id="connect4"></table></div></div>');
	for(var i = 0; i < 6; i += 1) {
		$('#connect4').append('<tr id="row_' + (5 - i).toString() + '" class="c4row"></tr>');
	}
	for(var i = 0; i < 7; i += 1) {
		$('.c4row').append('<td class="c4col col_' + i.toString() + '"></td>');
	}
	$('.c4col').click(make_move);
	$('.c4col').css('width', '100px');
	$('.c4col').css('height', '100px');
	$('.c4col').css('border', '1px solid black');
	$('.c4col').css('border-radius', '50%');
}

function connect_four() {
	userId = $('#userId').val().trim();
	if(userId == '') {
		alert("Must input User Id");
	} else {
		reset_board();
		socket = connect_socket(userId);
	}
}

function build_selector_str(row, col) {
	var rowS = 'row_' + row.toString();
	var colS = 'col_' + col.toString();
	return '#' + rowS + ' .' + colS;
}

function row_col(row, col) {
	return $(build_selector_str(row, col))
}

function connect_socket() {
	var socket = new WebSocket('ws://localhost:8080/game?userId=' + userId);
	socket.onmessage = function(event) {
		console.log(event.data);
		var board = JSON.parse(event.data);
		if(rematchSent && !board.GameOver) {
			reset_board();
			rematchSent = false;
			gameOver = false;
		}
		$('#turn_label').text("Current Turn: " + pieceColor[board.CurrentTurn]);
		$('#turn_label').css('color', pieceColor[board.CurrentTurn]);
		$('.c4col').css('background-color', null);
		for(var col = 0; col < board.Columns.length; col += 1) {
			for(var row = 0; row < board.Columns[col].length; row += 1) {
				var piece = board.Columns[col][row];
				var color = '';
				if(piece == 0) {
					color = 'red';
				} else {
					color = 'black';
				}
				row_col(row, col).css('background-color', color);
			}
		}
		if(board.GameOver && !gameOver) {
			gameOver = true;
			for(var i = 0; i < board.WinningPositions.length; i += 1) {
				row_col(board.WinningPositions[i].Row, board.WinningPositions[i].Col).css('border', '2px dashed green');
			}
			$('#sidebar').append('<input type="button" onclick="attempt_rematch()" value="Attempt Rematch">');
		}
	};

	socket.onclose = function(event) {
		alert("Socket Closed");
	};

	socket.onopen = function(event) {
	};

	return socket;
}


function make_move(event) {
	var classes = event.target.className.split(/\s+/);
	var colClicked = -1;
	for(var i = 0; i < classes.length; i += 1) {
		if(classes[i].startsWith('col_')) {
			colClicked = parseInt(classes[i].slice(4));
		}
	}
	if(colClicked < 0 || colClicked > 6) {
		return;
	}
	var turn = { Col: colClicked };
	socket.send(JSON.stringify(turn));
}

function attempt_rematch() {
	var rematch = {Rematch: true};
	rematchSent = true;
	socket.send(JSON.stringify(rematch));
}
