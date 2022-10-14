import React, { Component } from 'react';
import './App.css';
import tvimage from './snow.gif';

var client = new WebSocket('ws://127.0.0.1:8080');
client.binaryType = 'arraybuffer';

class App extends Component {
	state = {
		roll: "snow"
	  };
	  componentDidMount() {
	  }
		
	componentWillMount() {
		// setInterval(() => {
		// 	client.send('ciao') // sending any msg to start the server
		// 	console.log("pinging server...")
		// }, 1000);
  
		client.onopen = () => {
			console.log('WebSocket Client Connected');
		};
        client.onclose = () => {
            console.log("Closed...");
			// TODO: put back the snow
        }
		
		// https://stackoverflow.com/questions/39999367/how-do-i-reference-a-local-image-in-react
		// https://stackoverflow.com/a/64635519/436085
		client.onmessage = (msg) => {
			// console.dir(msg)
			var myImg = document.getElementById('myImg') as HTMLImageElement;
			var arrayBuffer = msg.data;
			var content = new Uint8Array(arrayBuffer);
	
			if (myImg) {
				myImg.src = URL.createObjectURL( new Blob([content.buffer], { type: 'image/png' } /* (1) */)  );
			}
		};
	}

	render() {
		
		return (
			<div className="App">
			<header className="App-header">
			
			<img src={tvimage} id="myImg" alt="myImg" />
			
			<p>
				Edit <code>src/App.tsx</code> and save to reload.
			  </p>
			  <a
				className="App-link"
				href="https://reactjs.org"
				target="_blank"
				rel="noopener noreferrer"
			  >
				Learn React
			  </a>
			  {this.state.roll }
			</header>
		  </div>
		);
	}
}


export default App;
