import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import User from './User/User'
import reportWebVitals from './reportWebVitals';
import { BrowserRouter, Routes, Route } from "react-router-dom";

export default function Root() {
	return (
		<BrowserRouter>
			<Routes>
				<Route path="/" element={<App />}/>
				<Route path="user" element={<User />}/>
			</Routes>
		</BrowserRouter>
	)
}

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<Root />)
//root.render(
  //React.StrictMode cause app to be rendered two times
  //<React.StrictMode>
    //<App />
  //</React.StrictMode>
//);

reportWebVitals();
