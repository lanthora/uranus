const { AppBar, Toolbar, Typography } = window['MaterialUI'];
const { Box, CircularProgress } = window['MaterialUI'];


const LOGIN_TITLE = "乌拉诺斯"

const root = ReactDOM.createRoot(document.getElementById('root'));

function Index() {
        const [userAlias, setUserAlias] = React.useState("");

        const updateUserInfo = () => {
                axios.post('/user/info', {}).then(function (response) {
                        setUserAlias(response.data.aliasName);
                });
        };

        React.useEffect(() => {
                document.title = LOGIN_TITLE;
                updateUserInfo();
        });

        return (
                <AppBar position="static">
                        <Toolbar variant="dense">
                                <Typography variant="h6" color="inherit" component="div">
                                        {userAlias}
                                </Typography>
                        </Toolbar>
                </AppBar>
        )
}

function Loading() {
        React.useEffect(() => {
                document.title = LOGIN_TITLE;
                axios.post('/user/alive', {}).then(function (response) {
                        root.render(<Index />);
                }).catch(function (error) {
                        if (error.response && error.response.status === 401) {
                                window.location.replace("/login");
                        }
                });
        });

        return (
                <div className="screen-center">
                        <Box sx={{ display: 'flex' }}>
                                <CircularProgress />
                        </Box>
                </div>
        )
}

root.render(<Loading />);



