import { useEffect } from 'react';
import { useRouter } from 'next/router';

const SnapRedirect = () => {
  const router = useRouter();
  const { token } = router.query;

  useEffect(() => {
    if (token) {
      window.snap.pay(token, {
        onSuccess: function (result) {
          router.push('/thankyou');
        },
        onPending: function (result) {
          alert('Waiting for your payment.');
        },
        onError: function (result) {
          alert('Payment failed.');
        },
        onClose: function () {
          alert('You closed the popup without finishing the payment.');
        },
      });
    }
  }, [token]);

  return <div>Redirecting to payment...</div>;
};

export default SnapRedirect;
