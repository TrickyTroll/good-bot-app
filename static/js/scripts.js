function playCast(x) {
    var audio = x.parentElement.querySelector('audio');
	var cast = x.parentElement.querySelector('asciinema-player');
    var playPause = audio.getAttribute('class');

    if (playPause === 'play') {
        audio.play();
		cast.play();
        audio.setAttribute('class', 'pause');

    } else {
        cast.pause();
		audio.pause();
        audio.setAttribute('class', 'play');
    }

    audio.onended = function () {
        cast.setAttribute('class', 'play');
    }
}
