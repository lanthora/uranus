const { AppBar, Toolbar, Typography } = window['MaterialUI'];

function UserInfoBar() {
        return (
                <AppBar position="static">
                        <Toolbar variant="dense">
                                <Typography variant="h6" color="inherit" component="div" id="alias">
                                        张三
                                </Typography>
                        </Toolbar>
                </AppBar>
        )
}

function Index() {
        return (
                <span>
                        <UserInfoBar />
                </span>
        )
}


const root = ReactDOM.createRoot(document.getElementById('root'));

axios.post('/user/current', {
}).then(function (response) {
        root.render(<Index />);
        console.log(response);
}).catch(function (error) {
        if (error.response && error.response.status === 401) {
                window.location.replace("/login");
        }
});


