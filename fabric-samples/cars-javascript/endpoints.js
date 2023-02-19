module.exports = function (app) {
	
    //read person asset
app.post('/readPersonAsset', (req, res) => {
	// Extract parameters from request body
	const { id } = req.body;
	readPersonAsset(contract, id);
	// Send response back to the client
	res.send(`Transaction completed.`);
});

//read car asset
app.post('/readCarAsset', (req, res) => {
	// Extract parameters from request body
	const { id } = req.body;
	readCarAsset(contract, id);
	// Send response back to the client
	res.send(`Transaction completed.`);
});

//add car malfunction
app.post('/addCarMalfunction', jsonParser, (req, res) => {
	// Extract parameters from request body
	console.log(req.body);
	const { id, description, repairPrice } = req.body;
	// Call addCarMalfunction function with parameters
	addCarMalfunction(contract, id, description, repairPrice);
	// Send response back to the client
	res.send(`Added car malfunction for car ${id}`);
});
}