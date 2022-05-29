const { Button, Snackbar, TextField, FormControl, Alert } = window['MaterialUI'];

const USERNAME = "用户名";
const PASSWORD = "密码";
const LOGIN = "登录";
const LOGIN_FAILED_MSG = "登录失败"
const LOGIN_FAILED_MSG_TIMEOUT = 3000


function Login() {

        const [open, setOpen] = React.useState(false);

        const username = React.useRef(null);
        const password = React.useRef(null);

        const openErrorNotify = () => {
                setOpen(true);
        };

        const closeErrorNotify = (event, reason) => {
                if (reason === 'clickaway') {
                        return;
                }
                setOpen(false);
        };

        function userLogin() {
                console.log(username.current.value);

                axios.post('/user/login', {
                        "username": username.current.value,
                        "password": password.current.value,
                }).then(function (response) {
                        if (response.status == 200) {
                                window.location
                                window.location.replace("/");
                        }
                }).catch(function (error) {
                        if (error.response && error.response.status === 401) {
                                openErrorNotify();
                        }
                });
        }


        return (
                <span>
                        <div className="login-center">
                                <FormControl>
                                        <TextField
                                                id="username"
                                                label={USERNAME}
                                                size="small"
                                                inputRef={username} />
                                        <TextField
                                                id="password"
                                                label={PASSWORD}
                                                size="small"
                                                inputRef={password}
                                                type="password"
                                                margin="dense" />
                                        <Button
                                                variant="contained"
                                                onClick={userLogin}>
                                                {LOGIN}
                                        </Button>
                                </FormControl>
                        </div>
                        <Snackbar
                                anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
                                open={open}
                                autoHideDuration={LOGIN_FAILED_MSG_TIMEOUT}
                                onClose={closeErrorNotify}>
                                <Alert
                                        onClose={closeErrorNotify}
                                        severity="error">
                                        {LOGIN_FAILED_MSG}
                                </Alert>
                        </Snackbar>
                </span>
        )
}

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<Login />);
